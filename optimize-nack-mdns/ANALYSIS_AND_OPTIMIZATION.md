# MediaMTX 서버 성능 분석 및 최적화 계획

## 1. 개요 (Executive Summary)
*   **증상:** 서버 구동 후 시간이 지남에 따라 CPU 사용량이 급증하고 메모리 소비가 줄어들지 않는 현상 발생.
*   **원인:** **WebRTC NACK(재전송 요청) 인터셉터**로 인한 과도한 메모리 할당 및 가비지 컬렉션(GC) 부하가 주원인입니다. `pion` 라이브러리는 패킷 재전송을 위해 송신 패킷을 버퍼에 저장하는데, 64개의 스트림과 다수의 연결이 동시에 작동하면서 이 버퍼가 막대한 메모리를 점유하고 있습니다.
*   **영향:** 실제 네트워크 트래픽 처리보다 메모리를 할당하고 해제하는 과정(GC)에서 CPU 자원이 소모되고 있어 성능 병목이 발생합니다.

## 2. 프로파일 상세 분석

### A. CPU 프로파일 분석 (`server_cpu_profile.prof`)
*   **시스템 오버헤드:** `syscall.Syscall6`(42%)가 가장 높으며, 이는 활발한 네트워크 I/O를 의미합니다.
*   **메모리 관리 오버헤드:** `runtime.mallocgc`(약 10%)와 `runtime.scanobject`가 상위권에 있습니다. 이는 애플리케이션이 메모리를 매우 빈번하게 할당하고 해제하고 있음을 나타내며, 이로 인해 **가비지 컬렉터(GC)**가 계속해서 실행되어 CPU를 점유하고 있습니다.
*   **WebRTC 로직:** `RTPWriterFunc.Write` 및 암호화 함수들이 실행되고 있지만, 이들은 결과적으로 메모리를 잡아먹는 버퍼에 데이터를 공급하는 역할을 하고 있습니다.

### B. 힙 메모리 프로파일 분석 (`server_heap_end.prof`)
*   **결정적 원인:** 전체 힙 메모리의 **72%**를 `github.com/pion/interceptor/internal/rtpbuffer.NewPacketFactoryCopy.func2` 함수가 점유하고 있습니다.
*   **의미:** 이 컴포넌트는 NACK Responder(응답기)의 일부로, 클라이언트가 패킷 손실을 신고했을 때 재전송하기 위해 **모든 발송 패킷의 복사본을 메모리에 저장**합니다.
*   **규모:** 64개의 고화질 스트림이 동시에 돌아가면서 이 "히스토리 버퍼"가 기하급수적으로 커졌고, 이것이 GC 스래싱(Thrashing)을 유발했습니다.

### C. 고루틴 프로파일
*   약 2,540개의 고루틴이 활성화되어 있으며, 대부분 `selectgo` 상태입니다. 이는 64개 스트림 x 시청자 수에 비례하는 수치로 보이며, GC가 실행될 때 관리해야 할 고루틴이 많아 스케줄링 오버헤드를 가중시킵니다.

## 3. 최적화 계획

### 1단계: NACK 비활성화를 통한 검증 (즉시 적용)
가장 빠르고 확실하게 성능을 확보하는 방법은 NACK 기능을 끄는 것입니다. 사내망(LAN)이나 네트워크 품질이 좋은 환경에서는 NACK 없이도 스트리밍에 문제가 없을 수 있습니다.

**조치:** `internal/protocols/webrtc/peer_connection.go` 파일에서 NACK 등록 코드를 주석 처리합니다.

### 2단계: 코드 레벨 최적화 (버퍼 크기 제한)
만약 패킷 손실이 잦아 NACK가 반드시 필요하다면, 기본 설정 대신 버퍼 크기를 줄여서 적용해야 합니다. `pion`의 기본 설정은 고밀도 서버에 비해 너무 공격적으로 메모리를 잡을 수 있습니다.

**`internal/protocols/webrtc/peer_connection.go` 수정 제안:**
`webrtc.ConfigureNack`을 사용하는 대신, 버퍼 사이즈를 제한하여 수동으로 인터셉터를 등록합니다.

```go
// 기존 코드 주석 처리 또는 삭제:
// err := webrtc.ConfigureNack(mediaEngine, interceptorRegistry)

// 수동 등록 추가 (import "github.com/pion/interceptor/pkg/nack" 필요):
// 버퍼 사이즈를 줄여서 (예: 1024 패킷) 메모리 사용량을 제한
responder, _ := nack.NewResponderInterceptor(nack.ResponderSize(1024))
interceptorRegistry.Add(responder)
```

### 3단계: 시스템 및 런타임 튜닝
*   **GOGC 조정:** 환경 변수 `GOGC=200`(기본값 100)으로 설정합니다. 이는 힙 메모리가 100% 증가할 때가 아니라 200% 증가할 때 GC를 수행하도록 하여, 메모리를 더 쓰는 대신 CPU(GC) 부하를 줄이는 전략입니다.
    *   실행 시: `export GOGC=200` 또는 `docker-compose.yml` 환경 변수에 추가.
*   **커널 버퍼:** 리눅스 커널의 UDP 버퍼를 늘려 시스템 콜 오버헤드를 줄입니다.
    ```bash
    sysctl -w net.core.rmem_max=26214400
    sysctl -w net.core.wmem_max=26214400
    ```

## 4. 실행 가이드
1.  **코드 수정:** `internal/protocols/webrtc/peer_connection.go`에서 `webrtc.ConfigureNack` 부분을 찾아서 주석 처리합니다.
2.  **빌드 및 배포:** 서버를 다시 빌드하여 배포합니다.
3.  **모니터링:** CPU 사용량이 안정화되는지 확인합니다.
4.  **조정:** 만약 화면 깨짐 현상(패킷 손실)이 발생하면, 2단계의 방법으로 버퍼 크기를 줄인 상태로 NACK를 다시 활성화합니다.
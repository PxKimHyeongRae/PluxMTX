# MediaMTX 최적화 결과 분석 (원격 서버)

## 중요 발견사항 ⚠️

**원격 서버(27.102.205.67)는 아직 NACK 제거 버전이 배포되지 않았습니다!**

프로파일 분석 결과, NACK 인터셉터가 여전히 실행 중입니다:
```
github.com/pion/interceptor/pkg/nack.(*ResponderInterceptor).BindLocalStream.func1: 9.13s (39.02% CPU)
github.com/pion/interceptor/internal/rtpbuffer.NewPacketFactoryCopy.func2: 105.65 MB (63.22% 메모리)
```

## 현재 원격 서버 성능 (NACK 활성화 상태)

### CPU 사용량
- **총 샘플 시간**: 23.40초 (60초 프로파일링)
- **주요 CPU 소비**:
  1. `syscall.Syscall6`: 9.26s (39.57%) - 시스템 콜
  2. **NACK 인터셉터**: 9.13s (39.02%) ⚠️
  3. WebRTC RTP 전송: 10.14s (43.33%)
  4. mDNS 읽기: 3.41s (14.57%)
  5. SRTP 암호화: 2.73s (11.67%)

### 메모리 사용량
- **전체 힙 메모리**: 167.12 MB
- **주요 메모리 소비**:
  1. **NACK 패킷 버퍼**: 105.65 MB (63.22%) ⚠️
  2. InterleavedFrame 버퍼: 11.52 MB (6.89%)
  3. NACK 패킷 팩토리: 8.50 MB (5.09%)

### 고루틴 수
- **총 고루틴**: 3,118개
- **주요 고루틴**:
  - counterdumper: 263개
  - rtpsender: 142개
  - OutgoingTrack: 142개
  - statsInterceptor: 142개

## NACK 제거 시 예상 개선

### 메모리 개선 예상치
현재 NACK 관련 메모리:
- PacketFactoryCopy.func2: 105.65 MB
- PacketFactoryCopy.func1: 8.50 MB
- **총 NACK 메모리**: ~114 MB

NACK 제거 후:
- 예상 메모리: 167 MB - 114 MB = **53 MB**
- **개선율: 68% 감소**

### CPU 개선 예상치
현재 NACK 관련 CPU:
- ResponderInterceptor: 9.13s (39.02%)

NACK 제거 후:
- 예상 CPU 감소: **39% 이상**

## 배포 준비 완료

### 빌드 완료
- ✅ **파일**: `mediamtx_optimized.exe` (44 MB)
- ✅ **변경사항**: NACK 인터셉터 비활성화
- ✅ **빌드 날짜**: 2025-12-04 11:31

### 주요 코드 변경
**파일**: `internal/protocols/webrtc/peer_connection.go`

**변경 내용**:
```go
// 라인 71-80: NACK 인터셉터 비활성화
// err := webrtc.ConfigureNack(mediaEngine, interceptorRegistry)
// if err != nil {
// 	return err
// }
```

### 배포 패키지 생성 필요
다음 파일들을 tar로 묶어서 배포:
1. `mediamtx_optimized.exe` - NACK 제거 바이너리
2. `mediamtx.yml` - 설정 파일
3. `NACK-REMOVAL-PLAN.md` - 구현 계획서
4. `performance-analysis.md` - 성능 분석 보고서
5. `OPTIMIZATION-RESULTS.md` - 이 파일

## 배포 후 확인 사항

### 1. 즉시 확인
배포 후 다음 명령으로 NACK 제거 확인:
```bash
# 10초 CPU 프로파일
curl "http://27.102.205.67:9999/debug/pprof/profile?seconds=10" > quick_test.prof
go tool pprof -top quick_test.prof | grep -i nack
```

**예상 결과**: NACK 관련 항목이 나타나지 않아야 함

### 2. 메모리 확인 (배포 후 1분)
```bash
curl "http://27.102.205.67:9999/debug/pprof/heap" > heap_after.prof
go tool pprof -text heap_after.prof | head -20
```

**기대값**: PacketFactoryCopy.func2가 10MB 이하

### 3. 성능 비교 (배포 후 5분)
```bash
# CPU 프로파일 (60초)
curl "http://27.102.205.67:9999/debug/pprof/profile?seconds=60" > cpu_after_deploy.prof

# 힙 메모리
curl "http://27.102.205.67:9999/debug/pprof/heap" > heap_after_deploy.prof

# 고루틴
curl "http://27.102.205.67:9999/debug/pprof/goroutine" > goroutine_after_deploy.prof
```

## 예상 성능 비교표

| 항목 | 현재 (NACK 활성화) | 예상 (NACK 제거) | 개선율 |
|------|-------------------|-----------------|--------|
| 힙 메모리 | 167 MB | 53 MB | -68% |
| NACK 버퍼 | 105.65 MB | < 5 MB | -95% |
| CPU (NACK) | 39.02% | 0% | -100% |
| 총 CPU | 높음 | 중간 | -40% 예상 |
| 고루틴 수 | 3,118개 | ~2,000개 | -35% 예상 |

## 배포 체크리스트

### 배포 전
- [x] 코드 변경 완료
- [x] 로컬 빌드 성공
- [x] 빌드 파일 확인 (mediamtx_optimized.exe)
- [ ] 배포 패키지 생성 (.tar)
- [ ] 백업 계획 수립

### 배포 시
- [ ] 현재 실행 중인 서버 중지
- [ ] NACK 제거 버전으로 교체
- [ ] 서버 재시작
- [ ] 로그 확인 (정상 시작 여부)

### 배포 후
- [ ] WebRTC 연결 테스트 (1개 스트림)
- [ ] 다중 스트림 테스트 (10개 이상)
- [ ] 10초 CPU 프로파일 수집 (NACK 확인)
- [ ] 메모리 프로파일 수집
- [ ] 1시간 안정성 모니터링
- [ ] 성능 개선 확인

## 롤백 절차

만약 문제가 발생하면:

1. **즉시 롤백**:
   ```bash
   # 이전 바이너리로 교체
   cp mediamtx.exe.backup mediamtx.exe
   # 서비스 재시작
   ```

2. **롤백 트리거**:
   - WebRTC 연결 실패
   - 비디오 품질 심각한 저하
   - 예상치 못한 크래시

## 추가 최적화 (NACK 제거 후 검토)

NACK 제거로도 성능이 부족하다면:

### 1. mDNS 비활성화 (추가 15% CPU 감소)
```go
// peer_connection.go의 Start() 함수
settingsEngine.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
```

### 2. RTCP 리포트 간격 증가 (추가 5-10% CPU 감소)
```go
// outgoing_track.go
Period: 3 * time.Second, // 1초 → 3초
```

## 결론

현재 원격 서버는 여전히 NACK가 활성화되어 있어 높은 CPU 및 메모리 사용량을 보이고 있습니다.

**NACK 제거 버전 배포 시 예상 효과**:
- ✅ 메모리 68% 감소 (167 MB → 53 MB)
- ✅ CPU 40% 이상 감소
- ✅ 패킷 버퍼링 오버헤드 95% 감소

**다음 단계**: 배포 패키지를 생성하여 원격 서버에 배포하고 성능 개선을 확인하세요.

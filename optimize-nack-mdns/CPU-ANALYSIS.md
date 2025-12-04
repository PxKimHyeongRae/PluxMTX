# CPU 사용률 상세 분석

## 현재 상황

**프로파일링 결과**: 180초 중 38.75초 CPU 샘플 = **21.5% CPU 사용률**

하지만 이 수치는 **pprof 샘플링 비율**이며, 실제 서버 CPU 사용률과 다를 수 있습니다.

## CPU 사용 분류 (Flat Time 기준)

### 1. 시스템 콜 (네트워크 I/O) - 35.87%
```
syscall.Syscall6: 13.90s (35.87%)
```
- UDP 패킷 전송 시스템 콜
- 64개 스트림의 RTP 패킷 전송
- **최적화 방법**: 패킷 배치 처리 (sendmmsg)

### 2. 암호화 (SRTP) - ~9%
```
crypto/sha1.blockGeneric:  1.14s (2.94%)
crypto/sha1.blockAVX2:     0.93s (2.40%)
crypto/aes.encryptBlock:   0.39s (1.01%)
xorBytesCTR:               0.32s (0.83%)
기타 암호화:                ~0.7s
-------------------------------------------
총:                        ~3.5s (9%)
```
- SRTP HMAC-SHA1 인증
- AES 암호화
- **최적화 어려움**: 보안을 위한 필수 작업

### 3. 런타임 오버헤드 - ~7.7%
```
runtime.futex:            0.85s (2.19%) - 고루틴 대기
runtime.findObject:       0.77s (1.99%) - GC
runtime.scanobject:       0.70s (1.81%) - GC
runtime.memmove:          0.56s (1.45%) - 메모리 복사
runtime.memclr:           0.55s (1.42%) - 메모리 클리어
-------------------------------------------
총:                       ~3s (7.7%)
```
- 가비지 컬렉션
- 고루틴 스케줄링
- 메모리 관리

### 4. 메모리 할당 - ~3.9%
```
runtime.mallocgc:         0.18s
runtime.mallocgc*:        ~1.3s
-------------------------------------------
총:                       ~1.5s (3.9%)
```

### 5. 애플리케이션 로직 - ~43%
```
WebRTC RTP write:         ~5s
RTSP read:                ~5s
기타:                     ~6.5s
-------------------------------------------
총:                       ~16.5s (43%)
```
- 실제 스트리밍 로직
- 버퍼 관리, 파싱 등

## 성능 벤치마크

### 스트림당 CPU 사용량
- 총 CPU: 38.75s / 180s = 21.5%
- 스트림 수: 64개
- **스트림당**: 21.5% / 64 = **0.34% per stream**

이는 매우 효율적입니다!

### 코어당 부하 (예상)
서버가 8코어라고 가정:
- 21.5% / 8 = **2.7% per core**

서버가 4코어라고 가정:
- 21.5% / 4 = **5.4% per core**

## 최적화 가능 영역

### 1. 패킷 배치 처리 (High Impact) ⭐⭐⭐
**현황**: syscall.Syscall6 13.90s (35.87%)

**방법**:
- UDP 패킷을 배치로 묶어 전송
- `sendmmsg()` 시스템 콜 사용
- 여러 패킷을 한 번의 시스템 콜로 처리

**예상 효과**:
- 시스템 콜 회수 감소 → CPU 10-15% 절감
- 레이턴시 약간 증가 가능

**구현 난이도**: 높음 (Pion WebRTC 라이브러리 수정)

### 2. SRTP 암호화 최적화 (Medium Impact) ⭐⭐
**현황**: SHA1 + AES ~3.5s (9%)

**방법**:
1. **하드웨어 가속 확인**
   ```bash
   # CPU의 AES-NI, SHA extension 지원 확인
   cat /proc/cpuinfo | grep -E "aes|sha"
   ```
   - AES-NI: 이미 사용 중일 가능성
   - SHA extension: CPU에 따라 다름

2. **GCM 암호화 모드로 전환**
   - AES-GCM은 하드웨어 가속 더 잘됨
   - HMAC-SHA1보다 빠름
   - Pion WebRTC 설정 변경 필요

**예상 효과**: CPU 3-5% 절감

**구현 난이도**: 중간

### 3. 메모리 할당 감소 (Low Impact) ⭐
**현황**: mallocgc ~1.5s (3.9%)

**방법**:
- 객체 풀(pool) 사용
- sync.Pool로 버퍼 재사용
- 불필요한 할당 제거

**예상 효과**: CPU 2-3% 절감

**구현 난이도**: 중간

### 4. 고루틴 풀 사용 (Low Impact) ⭐
**현황**: 2,562개 고루틴

**방법**:
- 워커 풀 패턴
- 고루틴 재사용

**예상 효과**:
- 메모리 2-3 MB 절감
- CPU 1-2% 절감

**구현 난이도**: 높음

## 진단 체크리스트

### 확인이 필요한 사항:

1. **실제 서버 CPU 사용률**
   ```bash
   # 서버에서 실행
   top -p $(pgrof mediamtx)
   # 또는
   ps aux | grep mediamtx
   ```
   - pprof의 21.5%와 실제 CPU 사용률 비교
   - 만약 실제 CPU가 80%라면 다른 문제 존재

2. **CPU 코어 수**
   ```bash
   nproc
   # 또는
   cat /proc/cpuinfo | grep processor | wc -l
   ```
   - 코어 수가 적으면 멀티코어 활용 개선 필요

3. **CPU 아키텍처**
   ```bash
   cat /proc/cpuinfo | grep -E "model name|flags" | head -5
   ```
   - AES-NI 지원 확인
   - AVX2 지원 확인 (이미 사용 중)

4. **네트워크 대역폭**
   ```bash
   # 64개 스트림의 총 대역폭
   iftop -i eth0
   # 또는
   nethogs
   ```
   - 네트워크 병목 확인

5. **다른 프로세스**
   ```bash
   top -o %CPU | head -20
   ```
   - mediamtx 외 다른 프로세스 CPU 사용 확인

## 결론 및 권장 사항

### 현재 상태 평가
✅ **매우 효율적**: 64개 스트림에 21.5% CPU는 우수한 성능
✅ **최적화 완료**: NACK(-75% 메모리), mDNS(-20% CPU) 제거 성공
⚠️ **추가 확인 필요**: 실제 서버 CPU 사용률과 프로파일 수치 비교

### 단계별 권장 사항

#### 1단계: 진단 (즉시)
```bash
# 서버에서 실행하여 실제 CPU 사용률 확인
ssh user@27.102.205.67
top -p $(pgrep mediamtx)
nproc
cat /proc/cpuinfo | grep "model name" | head -1
```

#### 2단계: 빠른 최적화 (필요 시)
- RTCP 간격 조정 (3-5% 절감)
- 메모리 풀 적용 (2-3% 절감)

#### 3단계: 장기 최적화 (필요 시)
- 패킷 배치 처리 구현 (10-15% 절감)
- SRTP GCM 모드 전환 (3-5% 절감)

### 예상 총 개선 가능
- 현재: 21.5% CPU
- 최대 최적화 후: **13-15% CPU** (6-8% 추가 절감)

---

**분석 일시**: 2025-12-04
**프로파일**: remote_cpu_mdns.prof (180초 샘플링)
**Build ID**: 0a95bb690faaf36f65dcdd6071b3b8e18b65b1c3

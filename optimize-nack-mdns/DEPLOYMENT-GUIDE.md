# MediaMTX NACK 최적화 버전 배포 가이드

## 패키지 정보

- **파일명**: `mediamtx-nack-optimized.tar.gz`
- **크기**: 22 MB
- **생성일**: 2025-12-04

### 포함 내용
```
mediamtx-nack-optimized/
├── mediamtx.exe                    # NACK 제거 최적화 바이너리
├── mediamtx.yml                    # 설정 파일
├── NACK-REMOVAL-PLAN.md           # 구현 계획서
├── OPTIMIZATION-RESULTS.md         # 성능 분석 결과
└── performance-analysis.md         # 초기 성능 분석
```

## 예상 성능 개선

| 항목 | 현재 | 예상 | 개선율 |
|------|------|------|--------|
| 힙 메모리 | 167 MB | 53 MB | **-68%** |
| NACK 버퍼 | 105.65 MB | < 5 MB | **-95%** |
| CPU (NACK) | 39.02% | 0% | **-100%** |
| 총 CPU | 높음 | 중간 | **-40%** |

## 배포 절차

### 1. 배포 전 준비

#### 1.1 현재 상태 백업
```bash
# 서버에서 실행
cd /path/to/mediamtx

# 현재 바이너리 백업
cp mediamtx mediamtx.backup.$(date +%Y%m%d)

# 설정 파일 백업
cp mediamtx.yml mediamtx.yml.backup.$(date +%Y%m%d)
```

#### 1.2 현재 성능 데이터 수집 (선택사항)
```bash
# 배포 전 프로파일 수집
curl "http://localhost:9999/debug/pprof/profile?seconds=30" > cpu_before.prof
curl "http://localhost:9999/debug/pprof/heap" > heap_before.prof
curl "http://localhost:9999/debug/pprof/goroutine" > goroutine_before.prof
```

### 2. 서버 중지

```bash
# systemd 사용 시
sudo systemctl stop mediamtx

# 또는 프로세스 직접 종료
pkill mediamtx

# 종료 확인
ps aux | grep mediamtx
```

### 3. 최적화 버전 배포

```bash
# 패키지 업로드 (로컬에서 서버로)
scp mediamtx-nack-optimized.tar.gz user@27.102.205.67:/tmp/

# 서버에서 실행
cd /tmp
tar -xzf mediamtx-nack-optimized.tar.gz

# 기존 설치 디렉토리로 이동
cd /path/to/mediamtx

# 바이너리 교체
cp /tmp/mediamtx-nack-optimized/mediamtx.exe ./mediamtx

# 실행 권한 부여
chmod +x mediamtx

# 설정 파일 확인 (필요 시 업데이트)
# cp /tmp/mediamtx-nack-optimized/mediamtx.yml ./
```

### 4. 서버 시작

```bash
# systemd 사용 시
sudo systemctl start mediamtx

# 또는 직접 실행 (테스트용)
./mediamtx &

# 시작 확인
ps aux | grep mediamtx
```

### 5. 로그 확인

```bash
# systemd 로그
sudo journalctl -u mediamtx -f

# 또는 파일 로그
tail -f mediamtx.log

# 정상 시작 메시지 예시:
# INF MediaMTX v0.0.0
# INF configuration loaded from mediamtx.yml
# INF [pprof] listener opened on :9999
# INF [WebRTC] listener opened on :8117
```

## 배포 후 검증

### 1. 기본 기능 테스트 (배포 직후)

#### 1.1 서비스 상태 확인
```bash
# HTTP API 체크
curl http://localhost:8119/v3/config/global

# pprof 엔드포인트 체크
curl http://localhost:9999/debug/pprof/
```

#### 1.2 WebRTC 연결 테스트
브라우저에서 1개 스트림 연결 후:
- 비디오 재생 확인
- 오디오 재생 확인 (있는 경우)
- 끊김 없는지 확인

### 2. NACK 제거 확인 (배포 후 1분)

```bash
# 빠른 CPU 프로파일 수집
curl "http://localhost:9999/debug/pprof/profile?seconds=10" > quick_test.prof

# NACK 관련 항목 검색
go tool pprof -top quick_test.prof | grep -i nack
```

**예상 결과**: NACK 관련 항목이 **나타나지 않아야 함**

**만약 NACK가 나타난다면**: 배포가 제대로 되지 않았으니 즉시 롤백

### 3. 메모리 확인 (배포 후 5분)

```bash
# 힙 메모리 프로파일
curl "http://localhost:9999/debug/pprof/heap" > heap_after.prof

# 상위 10개 항목 확인
go tool pprof -text heap_after.prof | head -20
```

**기대값**:
- PacketFactoryCopy.func2: **10 MB 이하** (이전: 105 MB)
- 전체 힙: **50-70 MB** (이전: 167 MB)

### 4. 다중 스트림 테스트 (배포 후 10분)

```bash
# 64개 스트림 연결 시도

# CPU 프로파일 (60초)
curl "http://localhost:9999/debug/pprof/profile?seconds=60" > cpu_after_full.prof

# 메모리 프로파일
curl "http://localhost:9999/debug/pprof/heap" > heap_after_full.prof

# 고루틴 프로파일
curl "http://localhost:9999/debug/pprof/goroutine" > goroutine_after_full.prof
```

### 5. 성능 비교 분석

```bash
# CPU 프로파일 비교
go tool pprof -top -cum cpu_after_full.prof | head -40

# 메모리 프로파일 비교
go tool pprof -text heap_after_full.prof | head -30

# 고루틴 수 확인
go tool pprof -text goroutine_after_full.prof | head -20
```

## 성공 기준

### 필수 기준 (모두 충족 필요)
- [ ] NACK 관련 코드가 프로파일에 나타나지 않음
- [ ] PacketFactoryCopy 메모리 < 10 MB
- [ ] 전체 힙 메모리 < 70 MB
- [ ] WebRTC 스트리밍 정상 작동
- [ ] 비디오 품질 저하 없음

### 목표 기준 (권장)
- [ ] 전체 힙 메모리 < 60 MB
- [ ] CPU 사용량 안정적 유지 (피크 없음)
- [ ] 패킷 손실률 < 0.5%

## 문제 발생 시 대응

### 롤백 트리거

다음 중 하나 발생 시 **즉시 롤백**:

1. **NACK가 여전히 프로파일에 나타남**
   - 배포가 제대로 되지 않음
   - 바이너리 교체 실패 가능성

2. **WebRTC 연결 실패**
   - 비디오 재생 안 됨
   - 연결 자체가 안 됨

3. **심각한 품질 저하**
   - 지속적인 끊김
   - 패킷 손실률 > 5%

4. **예상치 못한 크래시**
   - 서비스 다운
   - 패닉 에러

### 롤백 절차

```bash
# 1. 서비스 중지
sudo systemctl stop mediamtx

# 2. 백업 바이너리로 복원
cp mediamtx.backup.$(date +%Y%m%d) mediamtx

# 3. 서비스 재시작
sudo systemctl start mediamtx

# 4. 로그 확인
sudo journalctl -u mediamtx -f

# 5. 정상 작동 확인
curl http://localhost:8119/v3/config/global
```

### 부분 성공 시 (성능은 개선되었으나 목표 미달)

메모리는 개선되었지만 CPU가 여전히 높다면:

#### 옵션 A: mDNS 추가 비활성화
로컬에서 다음 수정 후 재배포:
```go
// internal/protocols/webrtc/peer_connection.go
settingsEngine.SetICEMulticastDNSMode(ice.MulticastDNSModeDisabled)
```

#### 옵션 B: RTCP 간격 조정
로컬에서 다음 수정 후 재배포:
```go
// internal/protocols/webrtc/outgoing_track.go
Period: 3 * time.Second, // 1초 → 3초
```

## 모니터링 계획

### 단기 모니터링 (배포 후 24시간)

```bash
# 1시간마다 실행
*/60 * * * * curl "http://localhost:9999/debug/pprof/heap" > /var/log/mediamtx/heap_$(date +\%Y\%m\%d_\%H\%M).prof

# CPU 모니터링
*/60 * * * * top -b -n 1 -p $(pgrep mediamtx) >> /var/log/mediamtx/cpu_usage.log
```

### 장기 모니터링 (1주일)

```bash
# 매일 자정에 프로파일 수집
0 0 * * * curl "http://localhost:9999/debug/pprof/profile?seconds=60" > /var/log/mediamtx/daily_cpu_$(date +\%Y\%m\%d).prof
0 0 * * * curl "http://localhost:9999/debug/pprof/heap" > /var/log/mediamtx/daily_heap_$(date +\%Y\%m\%d).prof
```

### 모니터링 지표

주기적으로 확인:
- **메모리**: `ps aux | grep mediamtx | awk '{print $6}'`
- **CPU**: `top -p $(pgrep mediamtx) -b -n 1 | grep mediamtx`
- **고루틴**: `curl -s http://localhost:9999/debug/pprof/goroutine?debug=1 | head -1`

## FAQ

### Q1: NACK 제거로 품질이 저하되지 않나요?
A: 안정적인 유선 네트워크 환경에서는 패킷 손실률이 매우 낮아(<0.1%) 영향이 거의 없습니다. 만약 Wi-Fi나 불안정한 네트워크에서 사용한다면 품질 저하가 발생할 수 있습니다.

### Q2: 롤백 시 데이터 손실이 있나요?
A: 없습니다. 바이너리만 교체하는 것이므로 설정이나 데이터는 그대로 유지됩니다.

### Q3: 배포 중 다운타임이 얼마나 되나요?
A: 서비스 중지 → 바이너리 교체 → 재시작까지 약 1-2분 소요됩니다.

### Q4: 메모리는 줄었는데 CPU가 여전히 높다면?
A: mDNS 비활성화 및 RTCP 간격 조정을 추가로 적용하세요 (위 "부분 성공 시" 참조).

### Q5: 프로파일 데이터를 어떻게 분석하나요?
A: 로컬로 다운로드 후 `go tool pprof` 명령어로 분석:
```bash
scp user@server:/var/log/mediamtx/heap_*.prof ./
go tool pprof -text heap_*.prof | head -30
```

## 연락처 및 지원

문제 발생 시:
1. 로그 파일 확인: `journalctl -u mediamtx -n 100`
2. 프로파일 데이터 수집: 위 명령어 참조
3. 롤백 실행: 위 절차 참조

## 체크리스트

### 배포 전
- [ ] 백업 완료 (바이너리 + 설정)
- [ ] 배포 전 프로파일 수집 (선택사항)
- [ ] 롤백 계획 숙지

### 배포 중
- [ ] 서비스 중지
- [ ] 바이너리 교체
- [ ] 권한 설정
- [ ] 서비스 시작
- [ ] 로그 확인

### 배포 후
- [ ] 기본 기능 테스트 (1분 내)
- [ ] NACK 제거 확인 (1분 내)
- [ ] 메모리 확인 (5분 후)
- [ ] 다중 스트림 테스트 (10분 후)
- [ ] 24시간 안정성 모니터링

---

**생성일**: 2025-12-04
**버전**: 1.0
**최적화 내용**: NACK 인터셉터 제거
docker build -t mediamtx:local .
docker save -o mediamtx.tar mediamtx:local
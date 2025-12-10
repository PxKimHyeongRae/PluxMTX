# MediaMTX with PTZ Support

**프로덕션 배포용 MediaMTX with Dynamic Dashboards & PTZ Control**

## 🎯 주요 기능

### 대시보드
- ✅ **WebRTC Dashboard** - 실시간 저지연 스트리밍 모니터링
- ✅ **HLS Dashboard** - 브라우저 호환 HTTP 스트리밍
- ✅ **PTZ Control** - 전용 카메라 제어 인터페이스

### PTZ 지원
- ✅ Hikvision ISAPI 통합
- ✅ 8방향 Pan/Tilt 제어
- ✅ Zoom In/Out
- ✅ 속도 조절 (10-100)
- ✅ 프리셋 관리

### 동적 로딩
- ✅ API 기반 스트림 목록 자동 로드
- ✅ 하드코딩 없음
- ✅ 실시간 설정 반영

## 🚀 빠른 배포

### 1. 환경 설정
```powershell
# 환경 변수 파일 생성
Copy-Item .env.example .env
```

### 2. 카메라 설정
`mediamtx.yml` 파일에 카메라 스트림 추가:
```yaml
paths:
  camera1:
    source: rtsp://user:pass@192.168.1.100:554/stream
    sourceOnDemand: yes
    rtspTransport: tcp
```

### 3. 배포 실행
```powershell
.\deploy.ps1
```

## 🌐 접속 URL

| 서비스 | URL |
|--------|-----|
| WebRTC 대시보드 | http://SERVER_IP:8889/dashboard |
| HLS 대시보드 | http://SERVER_IP:8889/dashboard-hls |
| PTZ 제어 | http://SERVER_IP:8889/ptz |
| API | http://SERVER_IP:9997/v3/paths/list |

## 📚 상세 문서

### 시작 가이드
- **[빠른 시작 가이드](docs/QUICK_GUIDE.md)** - 배포 및 설정 빠른 시작
- **[Docker 배포 가이드](docs/DOCKER_DEPLOYMENT.md)** - Docker를 이용한 배포 상세 가이드

### 기능 문서
- **[대시보드 기능](docs/DASHBOARD_README.md)** - WebRTC/HLS 대시보드 사용법
- **[PTZ API 명세](docs/PTZ_API.md)** - PTZ 제어 API 전체 명세
- **[PTZ 기술 가이드](docs/PTZ_TECHNICAL_GUIDE_KR.md)** - PTZ 제어 시스템 기술 문서

### 아키텍처
- **[WebRTC 아키텍처](docs/MEDIAMTX_WEBRTC_ARCHITECTURE.md)** - MediaMTX WebRTC 스트리밍 아키텍처
- **[PTZ 인터페이스 아키텍처](docs/PTZ_INTERFACE_ARCHITECTURE.md)** - PTZ 제어 인터페이스 설계
- **[ISAPI 심층 분석](docs/PTZ_ISAPI_DEEP_DIVE_KR.md)** - Hikvision ISAPI 프로토콜 분석

### 성능 및 테스트
- **[성능 최적화 보고서](docs/PERFORMANCE_OPTIMIZATION_REPORT.md)** - CPU/메모리 최적화 결과
- **[ONVIF 테스트 보고서](docs/ONVIF_TEST_REPORT.md)** - ONVIF PTZ 구현 테스트 결과
- **[Focus/Iris 구현](docs/FOCUS_IRIS.md)** - Focus 및 Iris 제어 구현 상세
- **[ONVIF Imaging 트러블슈팅](docs/ONVIF_IMAGING_TROUBLESHOOTING.md)** - ONVIF Focus/Iris 불완전 구현 원인 및 해결방안

## 📝 라이센스

MIT License

---

**상태**: ✅ Production Ready | **버전**: 1.0.0

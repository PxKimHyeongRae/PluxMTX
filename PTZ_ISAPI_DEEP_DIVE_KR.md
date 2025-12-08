# Hikvision ISAPI PTZ 심층 분석

## 목차
1. [ISAPI 문서는 어디서 찾나요?](#isapi-문서는-어디서-찾나요)
2. [PTZ 제어 방식 비교](#ptz-제어-방식-비교)
3. [각 방식의 XML 예시](#각-방식의-xml-예시)
4. [실전 예제 코드](#실전-예제-코드)
5. [카메라 기능 확인 방법](#카메라-기능-확인-방법)

---

## ISAPI 문서는 어디서 찾나요?

### 1. 공식 문서 출처

#### 🏢 Hikvision TPP (Technology Partner Program)
**URL**: https://tpp.hikvision.com

가장 공식적이고 최신 문서를 제공하는 곳입니다.

**접근 방법:**
1. TPP 포털 방문
2. "ISAPI & OTAP Developer Guide" 다운로드
3. 장치 시리즈/모델 번호로 검색
4. 해당 모델의 ISAPI 문서 다운로드

**필요한 문서:**
- `ISAPI General Application Developer Guide` - 전체 API 참조
- `ISAPI 2.0 PTZ Service Specification` - PTZ 전용 문서

#### 📄 직접 다운로드 가능한 문서

**ISAPI 2.0 PTZ Service PDF:**
```
https://download.catalogosicurezza.com/DOWNLOAD/Hikvision/Software/
Pacchetti per Sviluppo/05 ISAPI/HIKVISION ISAPI_2.0-PTZ Service.pdf
```

**General Application Developer Guide:**
```
https://download.isecj.jp/catalog/misc/isapi.pdf
```

**GitHub 미러:**
```
https://raw.githubusercontent.com/loozhengyuan/hikvision-sdk/master/resources/isapi.pdf
```

### 2. 문서 구조

ISAPI 문서는 일반적으로 다음과 같은 구조로 되어 있습니다:

```
ISAPI Developer Guide
├── 1. Introduction
├── 2. Authentication
├── 3. System Configuration
│   ├── Device Information
│   ├── Network Settings
│   └── ...
├── 15. PTZ Control (← 여기!)
│   ├── 15.1 PTZ Control Units
│   ├── 15.2 Get PTZ Capabilities
│   ├── 15.3 Continuous Movement
│   ├── 15.4 Momentary Movement
│   ├── 15.5 Relative Movement
│   ├── 15.6 Absolute Movement
│   ├── 15.7 Preset Management
│   └── 15.8 Auxiliary Controls
└── ...
```

### 3. 엔드포인트 찾는 방법

#### 방법 1: 문서 검색
```
1. PDF 문서 열기
2. "PTZCtrl" 검색
3. Section 15.x에서 원하는 기능 찾기
```

#### 방법 2: 카메라 자체 조회
카메라의 capabilities를 조회하면 지원하는 기능을 알 수 있습니다:

```bash
# 카메라가 지원하는 PTZ 기능 조회
curl -u admin:password \
  http://192.168.10.53/ISAPI/PTZCtrl/channels/1/capabilities
```

**응답 예시:**
```xml
<PTZChannelCapability>
  <ControlUnits>
    <ControlUnit>
      <id>continuous</id>
      <controlRange>
        <panRange><min>-100</min><max>100</max></panRange>
        <tiltRange><min>-100</min><max>100</max></tiltRange>
        <zoomRange><min>-100</min><max>100</max></zoomRange>
      </controlRange>
    </ControlUnit>
    <ControlUnit>
      <id>momentary</id>
      <!-- ... -->
    </ControlUnit>
    <ControlUnit>
      <id>relative</id>
      <!-- ... -->
    </ControlUnit>
  </ControlUnits>
</PTZChannelCapability>
```

#### 방법 3: 커뮤니티/오픈소스
- [IP Cam Talk 포럼](https://ipcamtalk.com/threads/figuring-out-hikvision-api-isapi.43619/)
- [ZoneMinder GitHub](https://github.com/ZoneMinder/zoneminder/blob/master/scripts/ZoneMinder/lib/ZoneMinder/Control/HikVision.pm)
- [Home Assistant 커뮤니티](https://community.home-assistant.io/t/hikvision-camera-ptz-control-workaround-without-onvif/180366)

### 4. 주요 PTZ 엔드포인트 목록

| 기능 | HTTP 메서드 | 엔드포인트 |
|------|------------|-----------|
| 연속 이동 | PUT | `/ISAPI/PTZCtrl/channels/1/continuous` |
| 순간 이동 | PUT | `/ISAPI/PTZCtrl/channels/1/momentary` |
| 상대 이동 | PUT | `/ISAPI/PTZCtrl/channels/1/relative` |
| 절대 이동 | PUT | `/ISAPI/PTZCtrl/channels/1/absolute` |
| 상태 조회 | GET | `/ISAPI/PTZCtrl/channels/1/status` |
| 기능 조회 | GET | `/ISAPI/PTZCtrl/channels/1/capabilities` |
| 프리셋 목록 | GET | `/ISAPI/PTZCtrl/channels/1/presets` |
| 프리셋 이동 | PUT | `/ISAPI/PTZCtrl/channels/1/presets/{id}/goto` |
| Zoom/Focus | PUT | `/ISAPI/PTZCtrl/channels/1/zoomFocus` |
| 보조 제어 | PUT | `/ISAPI/PTZCtrl/channels/1/auxcontrols/{id}` |

**참고**: `1`은 채널 번호입니다. 다채널 NVR의 경우 1, 2, 3...으로 변경됩니다.

---

## PTZ 제어 방식 비교

### 1. Continuous (연속 이동)

**개념**: 속도를 지정하여 계속 움직임. 명시적으로 정지할 때까지 계속 이동.

**엔드포인트**: `/ISAPI/PTZCtrl/channels/1/continuous`

**특징:**
- ✅ 속도 기반 제어 (`-100 ~ +100`)
- ✅ 정지 명령 전까지 계속 이동
- ✅ 부드러운 움직임
- ✅ 실시간 제어에 최적
- ❌ 정확한 위치 제어 어려움

**사용 사례:**
- 웹 UI의 방향 버튼 (누르는 동안 이동)
- 조이스틱 제어
- 실시간 추적

**동작 방식:**
```
1. Move(pan=50, tilt=0, zoom=0) 전송
   → 카메라가 우측으로 계속 회전 시작

2. (사용자가 버튼을 누르고 있는 동안 계속 회전)

3. Move(0, 0, 0) 전송 (또는 Stop() 호출)
   → 카메라가 즉시 정지
```

**실제 사용 예:**
```javascript
// 버튼을 누르는 순간
button.addEventListener('mousedown', () => {
  ptzMove(camera, 50, 0, 0);  // 우측으로 계속 회전
});

// 버튼을 떼는 순간
button.addEventListener('mouseup', () => {
  ptzStop(camera);  // 즉시 정지
});
```

### 2. Momentary (순간 이동)

**개념**: 지정한 시간(duration) 동안만 이동하고 자동으로 정지.

**엔드포인트**: `/ISAPI/PTZCtrl/channels/1/momentary`

**특징:**
- ✅ 시간 기반 제어
- ✅ 자동 정지 (별도 Stop 불필요)
- ✅ 정확한 시간 제어
- ❌ 실시간 조작감 떨어짐
- ❌ 버튼을 떼도 계속 움직임 (duration 끝날 때까지)

**사용 사례:**
- 프로그래밍된 패턴 이동
- 일정 각도만큼 회전
- 자동화된 스캔

**동작 방식:**
```
1. Move(pan=50, tilt=0, duration=2000ms) 전송
   → 카메라가 우측으로 2초간 회전

2. (2초 대기 - 자동으로 움직임)

3. 2초 후 자동으로 정지
```

**주의사항:**
```javascript
// 문제: 사용자가 버튼을 0.5초만 눌러도...
button.addEventListener('mousedown', () => {
  ptzMomentary(camera, 50, 0, 2000);  // 2초간 계속 회전!
});

// 버튼을 떼어도 2초가 지날 때까지 계속 회전
button.addEventListener('mouseup', () => {
  // 멈출 수 없음! (duration이 끝날 때까지)
});
```

**이것이 Continuous 대신 Momentary를 사용하지 않는 주요 이유입니다!**

### 3. Relative (상대 이동)

**개념**: 현재 위치에서 상대적으로 얼마나 이동할지 지정.

**엔드포인트**: `/ISAPI/PTZCtrl/channels/1/relative`

**특징:**
- ✅ 증분(increment) 기반 제어
- ✅ 현재 위치 조회 불필요
- ✅ 간단한 "조금 더 이동" 구현
- ❌ 누적 오차 발생 가능
- ❌ 정확한 위치 보장 어려움

**사용 사례:**
- "10도 더 회전" 같은 미세 조정
- 단계별 이동 (step-by-step)

**동작 방식:**
```
현재 위치: Pan=0°, Tilt=0°

1. Relative(pan=+10, tilt=+5) 전송
   → 새 위치: Pan=10°, Tilt=5°

2. Relative(pan=+10, tilt=+5) 다시 전송
   → 새 위치: Pan=20°, Tilt=10°

3. Relative(pan=-15, tilt=0) 전송
   → 새 위치: Pan=5°, Tilt=10°
```

**누적 오차 문제:**
```
이론:
10번의 +10° 이동 = +100° (예상)

실제:
10번의 +10° 이동 = +98.5° (실제 - 오차 누적)
```

### 4. Absolute (절대 이동)

**개념**: 정확한 좌표(각도)로 이동. GPS 좌표 같은 개념.

**엔드포인트**: `/ISAPI/PTZCtrl/channels/1/absolute`

**특징:**
- ✅ 정확한 위치 제어
- ✅ 오차 누적 없음
- ✅ 재현 가능한 위치
- ❌ 느린 응답 속도
- ❌ 단계적 움직임 (부드럽지 않음)

**사용 사례:**
- 정확한 각도 설정
- 프로그래밍된 투어
- 프리셋 대안

**동작 방식:**
```
1. Absolute(azimuth=1800, elevation=450) 전송
   → 카메라가 정확히 180.0°, 45.0° 위치로 이동

2. 현재 위치와 상관없이 항상 같은 위치로 이동
```

### 비교표

| 특징 | Continuous | Momentary | Relative | Absolute |
|------|-----------|-----------|----------|----------|
| **제어 방식** | 속도 | 속도 + 시간 | 증분 | 절대 좌표 |
| **정지 방법** | 명시적 Stop | 자동 (duration) | 자동 | 자동 |
| **정확도** | 낮음 | 중간 | 중간 | 높음 |
| **응답 속도** | 빠름 (~50ms) | 중간 (~100ms) | 느림 (~200ms) | 느림 (~500ms) |
| **부드러움** | 매우 부드러움 | 부드러움 | 단계적 | 단계적 |
| **실시간 제어** | 최적 | 부적합 | 부적합 | 부적합 |
| **UI 적합성** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐ |
| **자동화 적합성** | ⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |

### 왜 MediaMTX는 Continuous를 선택했나?

1. **사용자 경험 (UX)**
   - 버튼을 누르면 즉시 움직임
   - 버튼을 떼면 즉시 멈춤
   - 자연스러운 조작감

2. **실시간 응답**
   - 지연 시간 최소화 (~50ms)
   - 네트워크 지연에 강함

3. **모바일 최적화**
   - 터치 이벤트와 완벽 호환
   - "누르고 있기" 인터랙션 지원

4. **구현 단순성**
   - Stop = Move(0, 0, 0)
   - 복잡한 상태 관리 불필요

**Momentary를 사용하지 않는 이유:**
```javascript
// Momentary의 문제점
button.addEventListener('touchstart', () => {
  ptzMomentary(camera, 50, 0, 1000);  // 1초간 이동 시작
});

button.addEventListener('touchend', () => {
  // 문제: 사용자가 0.2초만 터치해도 1초간 계속 회전!
  // 멈출 수 없음!
});
```

---

## 각 방식의 XML 예시

### 1. Continuous (현재 사용 중)

**요청:**
```http
PUT http://192.168.10.53/ISAPI/PTZCtrl/channels/1/continuous
Content-Type: application/xml

<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>50</pan>      <!-- -100~100: 좌(음수)/우(양수) -->
    <tilt>30</tilt>    <!-- -100~100: 하(음수)/상(양수) -->
    <zoom>0</zoom>     <!-- -100~100: 축소(음수)/확대(양수) -->
</PTZData>
```

**정지:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>0</pan>
    <tilt>0</tilt>
    <zoom>0</zoom>
</PTZData>
```

**코드 구현:**
```go
// internal/ptz/hikvision.go:38-48
func (h *HikvisionPTZ) Move(pan, tilt, zoom int) error {
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>%d</pan>
    <tilt>%d</tilt>
    <zoom>%d</zoom>
</PTZData>`, pan, tilt, zoom)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/continuous", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}
```

### 2. Momentary

**요청:**
```http
PUT http://192.168.10.53/ISAPI/PTZCtrl/channels/1/momentary
Content-Type: application/xml

<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>50</pan>
    <tilt>0</tilt>
    <zoom>0</zoom>
    <Momentary>
        <duration>2000</duration>  <!-- 2초간 이동 (밀리초) -->
    </Momentary>
</PTZData>
```

**Go 구현 예시:**
```go
func (h *HikvisionPTZ) MoveMomentary(pan, tilt, zoom, durationMs int) error {
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>%d</pan>
    <tilt>%d</tilt>
    <zoom>%d</zoom>
    <Momentary>
        <duration>%d</duration>
    </Momentary>
</PTZData>`, pan, tilt, zoom, durationMs)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/momentary", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}
```

### 3. Relative

**요청:**
```http
PUT http://192.168.10.53/ISAPI/PTZCtrl/channels/1/relative
Content-Type: application/xml

<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <RelativeHigh>
        <elevation>10</elevation>   <!-- 현재 위치에서 +10 단위 상승 -->
        <azimuth>-15</azimuth>      <!-- 현재 위치에서 -15 단위 좌측 -->
        <absoluteZoom>5</absoluteZoom>  <!-- 줌 +5 단위 -->
    </RelativeHigh>
</PTZData>
```

**주의**: Relative는 `<pan>/<tilt>`가 아닌 `<azimuth>/<elevation>` 사용!

**Go 구현 예시:**
```go
func (h *HikvisionPTZ) MoveRelative(azimuth, elevation, zoom int) error {
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <RelativeHigh>
        <elevation>%d</elevation>
        <azimuth>%d</azimuth>
        <absoluteZoom>%d</absoluteZoom>
    </RelativeHigh>
</PTZData>`, elevation, azimuth, zoom)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/relative", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}
```

### 4. Absolute

**요청:**
```http
PUT http://192.168.10.53/ISAPI/PTZCtrl/channels/1/absolute
Content-Type: application/xml

<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <AbsoluteHigh>
        <elevation>450</elevation>      <!-- 45.0도 (x10) -->
        <azimuth>1800</azimuth>         <!-- 180.0도 (x10) -->
        <absoluteZoom>50</absoluteZoom> <!-- 줌 레벨 50 -->
    </AbsoluteHigh>
</PTZData>
```

**각도 변환:**
- XML 값 = 실제 각도 × 10
- 예: 45.5° → 455
- 예: 180.0° → 1800

**Go 구현 예시:**
```go
func (h *HikvisionPTZ) MoveAbsolute(azimuthDegrees, elevationDegrees float64, zoom int) error {
	// 각도를 x10으로 변환
	azimuth := int(azimuthDegrees * 10)
	elevation := int(elevationDegrees * 10)

	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <AbsoluteHigh>
        <elevation>%d</elevation>
        <azimuth>%d</azimuth>
        <absoluteZoom>%d</absoluteZoom>
    </AbsoluteHigh>
</PTZData>`, elevation, azimuth, zoom)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/absolute", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}
```

---

## 실전 예제 코드

### 시나리오 1: 웹 UI 버튼 제어 (Continuous 사용)

**JavaScript (프론트엔드):**
```javascript
let isPTZActive = false;

// 상 버튼
const upButton = document.getElementById('ptz-up');

upButton.addEventListener('mousedown', () => {
  isPTZActive = true;
  ptzMove('CCTV-TEST1', 0, 40, 0);  // 위로 이동 시작
});

upButton.addEventListener('mouseup', () => {
  isPTZActive = false;
  ptzStop('CCTV-TEST1');  // 즉시 정지
});

// 터치 디바이스 지원
upButton.addEventListener('touchstart', (e) => {
  e.preventDefault();
  isPTZActive = true;
  ptzMove('CCTV-TEST1', 0, 40, 0);
});

upButton.addEventListener('touchend', (e) => {
  e.preventDefault();
  isPTZActive = false;
  ptzStop('CCTV-TEST1');
});
```

**Go (백엔드):**
```go
// internal/ptz/hikvision.go
func (h *HikvisionPTZ) Move(pan, tilt, zoom int) error {
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>%d</pan>
    <tilt>%d</tilt>
    <zoom>%d</zoom>
</PTZData>`, pan, tilt, zoom)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/continuous", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}

func (h *HikvisionPTZ) Stop() error {
	return h.Move(0, 0, 0)
}
```

### 시나리오 2: 자동 패턴 스캔 (Momentary 사용)

**Go 코드:**
```go
func (h *HikvisionPTZ) ScanPattern() error {
	// 좌측으로 2초
	h.MoveMomentary(-50, 0, 0, 2000)
	time.Sleep(2 * time.Second)

	// 우측으로 4초
	h.MoveMomentary(50, 0, 0, 4000)
	time.Sleep(4 * time.Second)

	// 원점 복귀
	h.MoveMomentary(-50, 0, 0, 2000)
	time.Sleep(2 * time.Second)

	return nil
}

func (h *HikvisionPTZ) MoveMomentary(pan, tilt, zoom, durationMs int) error {
	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <pan>%d</pan>
    <tilt>%d</tilt>
    <zoom>%d</zoom>
    <Momentary>
        <duration>%d</duration>
    </Momentary>
</PTZData>`, pan, tilt, zoom, durationMs)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/momentary", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}
```

### 시나리오 3: 정확한 각도로 이동 (Absolute 사용)

**Go 코드:**
```go
// 주요 위치 정의
type CameraPosition struct {
	Name      string
	Azimuth   float64  // 수평 각도 (0-360)
	Elevation float64  // 수직 각도 (-90 ~ +90)
	Zoom      int      // 줌 레벨
}

var presetPositions = []CameraPosition{
	{"Front Door", 0.0, 0.0, 10},
	{"Parking Lot", 90.0, -15.0, 20},
	{"Back Yard", 180.0, -5.0, 15},
	{"Side Entrance", 270.0, 0.0, 10},
}

func (h *HikvisionPTZ) GoToPosition(pos CameraPosition) error {
	azimuth := int(pos.Azimuth * 10)      // 0° → 0, 90° → 900
	elevation := int(pos.Elevation * 10)  // -15° → -150

	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <AbsoluteHigh>
        <elevation>%d</elevation>
        <azimuth>%d</azimuth>
        <absoluteZoom>%d</absoluteZoom>
    </AbsoluteHigh>
</PTZData>`, elevation, azimuth, pos.Zoom)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/absolute", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}

// 사용 예시
func main() {
	ptz := NewHikvisionPTZ("192.168.10.53", "admin", "password")

	// Front Door 위치로 이동
	ptz.GoToPosition(presetPositions[0])
	time.Sleep(3 * time.Second)

	// Parking Lot 위치로 이동
	ptz.GoToPosition(presetPositions[1])
}
```

### 시나리오 4: 미세 조정 (Relative 사용)

**Go 코드:**
```go
// 현재 위치에서 조금씩 이동
func (h *HikvisionPTZ) Nudge(direction string) error {
	var azimuth, elevation int

	switch direction {
	case "up":
		elevation = 5      // 위로 5 단위
	case "down":
		elevation = -5     // 아래로 5 단위
	case "left":
		azimuth = -5       // 좌로 5 단위
	case "right":
		azimuth = 5        // 우로 5 단위
	}

	xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<PTZData>
    <RelativeHigh>
        <elevation>%d</elevation>
        <azimuth>%d</azimuth>
        <absoluteZoom>0</absoluteZoom>
    </RelativeHigh>
</PTZData>`, elevation, azimuth)

	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/relative", h.Host)
	return h.sendRequest("PUT", url, xmlData)
}

// 사용 예시
func main() {
	ptz := NewHikvisionPTZ("192.168.10.53", "admin", "password")

	// 위로 조금
	ptz.Nudge("up")
	time.Sleep(500 * time.Millisecond)

	// 우측으로 조금
	ptz.Nudge("right")
	time.Sleep(500 * time.Millisecond)

	// 아래로 조금
	ptz.Nudge("down")
}
```

---

## 카메라 기능 확인 방법

### 1. Capabilities 조회

모든 카메라가 모든 PTZ 모드를 지원하는 것은 아닙니다. 먼저 확인이 필요합니다.

**요청:**
```bash
curl -u admin:password \
  http://192.168.10.53/ISAPI/PTZCtrl/channels/1/capabilities
```

**응답 예시:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<PTZChannelCapability version="2.0">
  <ControlUnits>
    <!-- Continuous 지원 -->
    <ControlUnit>
      <id>continuous</id>
      <controlRange>
        <panRange>
          <min>-100</min>
          <max>100</max>
        </panRange>
        <tiltRange>
          <min>-100</min>
          <max>100</max>
        </tiltRange>
        <zoomRange>
          <min>-100</min>
          <max>100</max>
        </zoomRange>
      </controlRange>
    </ControlUnit>

    <!-- Momentary 지원 -->
    <ControlUnit>
      <id>momentary</id>
      <controlRange>
        <panRange>
          <min>-100</min>
          <max>100</max>
        </panRange>
        <tiltRange>
          <min>-100</min>
          <max>100</max>
        </tiltRange>
        <zoomRange>
          <min>-100</min>
          <max>100</max>
        </zoomRange>
        <durationRange>
          <min>100</min>       <!-- 최소 100ms -->
          <max>10000</max>     <!-- 최대 10초 -->
        </durationRange>
      </controlRange>
    </ControlUnit>

    <!-- Relative 지원 -->
    <ControlUnit>
      <id>relative</id>
      <controlRange>
        <elevationRange>
          <min>-3600</min>     <!-- -360도 -->
          <max>3600</max>      <!-- +360도 -->
        </elevationRange>
        <azimuthRange>
          <min>-3600</min>
          <max>3600</max>
        </azimuthRange>
      </controlRange>
    </ControlUnit>

    <!-- Absolute 지원 -->
    <ControlUnit>
      <id>absolute</id>
      <controlRange>
        <elevationRange>
          <min>-900</min>      <!-- -90도 -->
          <max>900</max>       <!-- +90도 -->
        </elevationRange>
        <azimuthRange>
          <min>0</min>         <!-- 0도 -->
          <max>3600</max>      <!-- 360도 -->
        </azimuthRange>
      </controlRange>
    </ControlUnit>
  </ControlUnits>

  <!-- 프리셋 지원 -->
  <SupportedPresetNum>300</SupportedPresetNum>

  <!-- 기타 기능 -->
  <SupportAuxControl>true</SupportAuxControl>
  <SupportPatternScan>true</SupportPatternScan>
</PTZChannelCapability>
```

### 2. Go 코드로 기능 확인

**구현:**
```go
type PTZCapabilities struct {
	SupportsContinuous bool
	SupportsMomentary  bool
	SupportsRelative   bool
	SupportsAbsolute   bool
	MaxPresets         int
}

func (h *HikvisionPTZ) GetCapabilities() (*PTZCapabilities, error) {
	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/capabilities", h.Host)
	xmlData, err := h.sendGetRequest(url)
	if err != nil {
		return nil, err
	}

	caps := &PTZCapabilities{}

	// 간단한 문자열 검색 (실제로는 XML 파싱 권장)
	caps.SupportsContinuous = strings.Contains(xmlData, "<id>continuous</id>")
	caps.SupportsMomentary = strings.Contains(xmlData, "<id>momentary</id>")
	caps.SupportsRelative = strings.Contains(xmlData, "<id>relative</id>")
	caps.SupportsAbsolute = strings.Contains(xmlData, "<id>absolute</id>")

	return caps, nil
}

// 사용 예시
func main() {
	ptz := NewHikvisionPTZ("192.168.10.53", "admin", "password")

	caps, err := ptz.GetCapabilities()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Continuous: %v\n", caps.SupportsContinuous)
	fmt.Printf("Momentary: %v\n", caps.SupportsMomentary)
	fmt.Printf("Relative: %v\n", caps.SupportsRelative)
	fmt.Printf("Absolute: %v\n", caps.SupportsAbsolute)
}
```

### 3. 현재 상태 조회

**요청:**
```bash
curl -u admin:password \
  http://192.168.10.53/ISAPI/PTZCtrl/channels/1/status
```

**응답:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<PTZStatus version="2.0">
  <AbsoluteHigh>
    <azimuth>1800</azimuth>        <!-- 현재 180.0도 -->
    <elevation>450</elevation>     <!-- 현재 45.0도 -->
    <absoluteZoom>25</absoluteZoom><!-- 현재 줌 레벨 25 -->
  </AbsoluteHigh>
  <PTZUtcTime>2025-12-08T10:30:45Z</PTZUtcTime>
</PTZStatus>
```

**Go 구현:**
```go
type PTZStatus struct {
	Azimuth   float64  // 수평 각도
	Elevation float64  // 수직 각도
	Zoom      int      // 줌 레벨
}

func (h *HikvisionPTZ) GetCurrentPosition() (*PTZStatus, error) {
	url := fmt.Sprintf("http://%s/ISAPI/PTZCtrl/channels/1/status", h.Host)
	xmlData, err := h.sendGetRequest(url)
	if err != nil {
		return nil, err
	}

	// 간단한 정규식 파싱 (실제로는 XML 파서 권장)
	status := &PTZStatus{}

	// <azimuth>1800</azimuth> → 180.0
	if match := regexp.MustCompile(`<azimuth>(\d+)</azimuth>`).FindStringSubmatch(xmlData); len(match) > 1 {
		val, _ := strconv.Atoi(match[1])
		status.Azimuth = float64(val) / 10.0
	}

	// <elevation>450</elevation> → 45.0
	if match := regexp.MustCompile(`<elevation>(-?\d+)</elevation>`).FindStringSubmatch(xmlData); len(match) > 1 {
		val, _ := strconv.Atoi(match[1])
		status.Elevation = float64(val) / 10.0
	}

	// <absoluteZoom>25</absoluteZoom> → 25
	if match := regexp.MustCompile(`<absoluteZoom>(\d+)</absoluteZoom>`).FindStringSubmatch(xmlData); len(match) > 1 {
		status.Zoom, _ = strconv.Atoi(match[1])
	}

	return status, nil
}

// 사용 예시
func main() {
	ptz := NewHikvisionPTZ("192.168.10.53", "admin", "password")

	status, err := ptz.GetCurrentPosition()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("현재 위치: Azimuth=%.1f°, Elevation=%.1f°, Zoom=%d\n",
		status.Azimuth, status.Elevation, status.Zoom)
}
```

---

## 정리

### ISAPI 문서 찾기
1. **공식**: https://tpp.hikvision.com (TPP 포털)
2. **커뮤니티**: IP Cam Talk, ZoneMinder GitHub
3. **카메라 자체**: `/ISAPI/PTZCtrl/channels/1/capabilities`

### PTZ 제어 방식 선택 가이드

| 사용 사례 | 권장 방식 | 이유 |
|----------|---------|------|
| 웹 UI 버튼 | **Continuous** | 실시간 제어, 부드러운 움직임 |
| 모바일 터치 | **Continuous** | 터치 이벤트와 완벽 호환 |
| 자동 패턴 | **Momentary** | 시간 기반 자동 정지 |
| 정확한 위치 | **Absolute** | 좌표 기반 재현 가능 |
| 미세 조정 | **Relative** | 현재 위치에서 증분 이동 |
| 프리셋 | **GotoPreset** | 가장 빠르고 정확 |

### MediaMTX가 Continuous를 선택한 핵심 이유

```javascript
// ✅ Continuous: 자연스러운 UX
button.onmousedown → move(50, 0, 0)  // 즉시 이동
button.onmouseup   → move(0, 0, 0)   // 즉시 정지

// ❌ Momentary: 불편한 UX
button.onmousedown → move(50, 0, 1000)  // 1초간 이동 시작
button.onmouseup   → (멈출 수 없음!)   // 1초 끝날 때까지 계속 회전
```

---

## 참고 자료

- [Hikvision TPP Portal](https://tpp.hikvision.com)
- [ISAPI 2.0 PTZ Service PDF](https://download.catalogosicurezza.com/DOWNLOAD/Hikvision/Software/Pacchetti per Sviluppo/05   ISAPI/HIKVISION ISAPI_2.0-PTZ Service.pdf)
- [ISAPI General Application Developer Guide](https://download.isecj.jp/catalog/misc/isapi.pdf)
- [IP Cam Talk - Figuring out Hikvision API](https://ipcamtalk.com/threads/figuring-out-hikvision-api-isapi.43619/)
- [ZoneMinder Hikvision Control](https://github.com/ZoneMinder/zoneminder/blob/master/scripts/ZoneMinder/lib/ZoneMinder/Control/HikVision.pm)
- [Home Assistant Hikvision PTZ](https://community.home-assistant.io/t/hikvision-camera-ptz-control-workaround-without-onvif/180366)

---

**작성일**: 2025-12-08
**버전**: 1.0
**작성자**: MediaMTX PTZ 개발팀

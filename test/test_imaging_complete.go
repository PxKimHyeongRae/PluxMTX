package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/use-go/onvif"
	onvif_imaging "github.com/use-go/onvif/Imaging"
	onvif_media "github.com/use-go/onvif/media"
	onvif_ptz "github.com/use-go/onvif/ptz"
	"github.com/use-go/onvif/xsd"
	xsd_onvif "github.com/use-go/onvif/xsd/onvif"
)

func main() {
	host := "14.51.233.129"
	port := 10082
	username := "admin"
	password := "pluxity123!@#"

	fmt.Printf("=== ONVIF Imaging Service 완전 테스트 ===\n")
	fmt.Printf("=== 놓친 테스트 항목 모두 실행 ===\n\n")

	// ONVIF 장치 연결
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    fmt.Sprintf("%s:%d", host, port),
		Username: username,
		Password: password,
	})
	if err != nil {
		fmt.Printf("❌ 연결 실패: %v\n", err)
		return
	}

	fmt.Println("✅ ONVIF 장치 연결 성공")

	// Get media profiles
	getProfilesReq := onvif_media.GetProfiles{}
	profilesResp, err := dev.CallMethod(getProfilesReq)
	if err != nil {
		fmt.Printf("❌ GetProfiles 실패: %v\n", err)
		return
	}

	body, _ := io.ReadAll(profilesResp.Body)
	profilesResp.Body.Close()

	var envelope struct {
		Body struct {
			GetProfilesResponse struct {
				Profiles []struct {
					Token                    string `xml:"token,attr"`
					Name                     string
					VideoSourceConfiguration struct {
						SourceToken string
					}
				}
			}
		}
	}

	if err := xml.Unmarshal(body, &envelope); err != nil {
		fmt.Printf("❌ 프로파일 파싱 실패: %v\n", err)
		return
	}

	if len(envelope.Body.GetProfilesResponse.Profiles) == 0 {
		fmt.Printf("❌ 프로파일을 찾을 수 없습니다\n")
		return
	}

	profile := envelope.Body.GetProfilesResponse.Profiles[0]
	videoSourceToken := xsd_onvif.ReferenceToken(profile.VideoSourceConfiguration.SourceToken)
	profileToken := xsd_onvif.ReferenceToken(profile.Token)

	fmt.Printf("✅ VideoSource Token: %s\n", videoSourceToken)
	fmt.Printf("✅ Profile Token: %s\n\n", profileToken)

	// ========================================
	// 테스트 1: GetServiceCapabilities (미테스트였음)
	// ========================================
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 1: Imaging.GetServiceCapabilities ===")
	fmt.Println("=== (Imaging 서비스가 지원하는 기능 확인) ===")
	testGetServiceCapabilities(dev)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 2: GetStatus (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 2: Imaging.GetStatus ===")
	fmt.Println("=== (Focus/Iris 현재 상태 조회) ===")
	testImagingGetStatus(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 3: Move - Absolute Focus (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 3: Imaging.Move - Absolute Focus ===")
	fmt.Println("=== (절대 위치로 Focus 이동) ===")
	testAbsoluteFocusMove(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 4: Move - Relative Focus (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 4: Imaging.Move - Relative Focus ===")
	fmt.Println("=== (상대 거리만큼 Focus 이동) ===")
	testRelativeFocusMove(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 5: GetPresets (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 5: Imaging.GetPresets ===")
	fmt.Println("=== (Focus 프리셋 목록 조회) ===")
	testImagingGetPresets(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 6: GetCurrentPreset (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 6: Imaging.GetCurrentPreset ===")
	fmt.Println("=== (현재 Focus 프리셋 조회) ===")
	testImagingGetCurrentPreset(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 7: SetImagingSettings - Focus.AutoFocusMode (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 7: SetImagingSettings - AutoFocusMode ===")
	fmt.Println("=== (AutoFocus 모드 전환: MANUAL) ===")
	testSetAutoFocusMode(dev, videoSourceToken, "MANUAL")
	time.Sleep(1 * time.Second)

	fmt.Println("\n=== 테스트 7-2: SetImagingSettings - AutoFocusMode ===")
	fmt.Println("=== (AutoFocus 모드 전환: AUTO) ===")
	testSetAutoFocusMode(dev, videoSourceToken, "AUTO")
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 8: 추가 PTZ Auxiliary Commands (미테스트였음)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 8: 추가 PTZ Auxiliary Commands ===")
	testAdditionalAuxCommands(dev, profileToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 9: Continuous Focus 재확인 (다른 속도값)
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 9: Continuous Focus (정수 Speed 값) ===")
	testContinuousFocusWithIntSpeed(dev, videoSourceToken)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 모든 테스트 완료 ===")
}

// 테스트 1: GetServiceCapabilities
func testGetServiceCapabilities(dev *onvif.Device) {
	req := onvif_imaging.GetServiceCapabilities{}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("응답 코드: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ GetServiceCapabilities 성공!")
		fmt.Printf("응답:\n%s\n", string(body))

		// 주요 키워드 검색
		keywords := []string{"ImageStabilization", "Presets", "Focus", "Iris"}
		for _, kw := range keywords {
			if strings.Contains(string(body), kw) {
				fmt.Printf("  🔍 '%s' 지원 확인\n", kw)
			}
		}
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
}

// 테스트 2: GetStatus
func testImagingGetStatus(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.GetStatus{
		VideoSourceToken: token,
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("응답 코드: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ GetStatus 성공!")
		fmt.Printf("응답:\n%s\n", string(body))

		// Focus 상태 파싱
		var statusEnvelope struct {
			Body struct {
				GetStatusResponse struct {
					Status struct {
						FocusStatus20 struct {
							Position   float64 `xml:"Position"`
							MoveStatus string  `xml:"MoveStatus"`
							Error      string  `xml:"Error"`
						} `xml:"FocusStatus20"`
					} `xml:"Status"`
				} `xml:"GetStatusResponse"`
			} `xml:"Body"`
		}

		if err := xml.Unmarshal(body, &statusEnvelope); err == nil {
			status := statusEnvelope.Body.GetStatusResponse.Status.FocusStatus20
			fmt.Printf("\n📍 Focus 상태:\n")
			fmt.Printf("   Position: %.4f\n", status.Position)
			fmt.Printf("   MoveStatus: %s\n", status.MoveStatus)
			if status.Error != "" {
				fmt.Printf("   Error: %s\n", status.Error)
			}
		}
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
}

// 테스트 3: Absolute Focus Move
func testAbsoluteFocusMove(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	positions := []float64{0.0, 0.5, 1.0}

	for _, pos := range positions {
		fmt.Printf("\n--- Absolute Focus Position: %.1f ---\n", pos)

		req := onvif_imaging.Move{
			VideoSourceToken: token,
			Focus: xsd_onvif.FocusMove{
				Absolute: xsd_onvif.AbsoluteFocus{
					Position: xsd.Float(pos),
					Speed:    xsd.Float(1.0),
				},
			},
		}

		resp, err := dev.CallMethod(req)
		if err != nil {
			fmt.Printf("❌ 에러: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("응답 코드: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			fmt.Printf("✅ Absolute Focus Move 성공! (Position: %.1f)\n", pos)
		} else {
			fmt.Printf("❌ 실패: %s\n", string(body))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// 테스트 4: Relative Focus Move
func testRelativeFocusMove(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	distances := []float64{0.1, -0.1, 0.5, -0.5}

	for _, dist := range distances {
		fmt.Printf("\n--- Relative Focus Distance: %.1f ---\n", dist)

		req := onvif_imaging.Move{
			VideoSourceToken: token,
			Focus: xsd_onvif.FocusMove{
				Relative: xsd_onvif.RelativeFocus{
					Distance: xsd.Float(dist),
					Speed:    xsd.Float(1.0),
				},
			},
		}

		resp, err := dev.CallMethod(req)
		if err != nil {
			fmt.Printf("❌ 에러: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("응답 코드: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			fmt.Printf("✅ Relative Focus Move 성공! (Distance: %.1f)\n", dist)
		} else {
			fmt.Printf("❌ 실패: %s\n", string(body))
		}

		time.Sleep(500 * time.Millisecond)
	}
}

// 테스트 5: GetPresets
func testImagingGetPresets(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.GetPresets{
		VideoSourceToken: token,
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("응답 코드: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ GetPresets 성공!")
		fmt.Printf("응답:\n%s\n", string(body))
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
}

// 테스트 6: GetCurrentPreset
func testImagingGetCurrentPreset(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.GetCurrentPreset{
		VideoSourceToken: token,
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("응답 코드: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ GetCurrentPreset 성공!")
		fmt.Printf("응답:\n%s\n", string(body))
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
}

// 테스트 7: SetImagingSettings - AutoFocusMode
func testSetAutoFocusMode(dev *onvif.Device, token xsd_onvif.ReferenceToken, mode string) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Focus: xsd_onvif.FocusConfiguration20{
				AutoFocusMode: xsd_onvif.AutoFocusMode(mode),
			},
		},
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("응답 코드: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Printf("✅ AutoFocusMode='%s' 설정 성공!\n", mode)
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
}

// 테스트 8: 추가 Auxiliary Commands
func testAdditionalAuxCommands(dev *onvif.Device, profileToken xsd_onvif.ReferenceToken) {
	commands := []string{
		"tt:FocusNear",
		"tt:FocusFar",
		"tt:AutoFocus",
		"FocusNear",
		"FocusFar",
		"Focus+",
		"Focus-",
		"AutoFocusOn",
		"AutoFocusOff",
	}

	for _, cmd := range commands {
		fmt.Printf("\n--- AuxiliaryCommand: %s ---\n", cmd)

		req := onvif_ptz.SendAuxiliaryCommand{
			ProfileToken:  profileToken,
			AuxiliaryData: xsd_onvif.AuxiliaryData(cmd),
		}

		resp, err := dev.CallMethod(req)
		if err != nil {
			fmt.Printf("❌ 에러: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("응답 코드: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			fmt.Printf("✅ '%s' 명령 성공!\n", cmd)
		} else {
			// 에러 메시지에서 핵심만 추출
			if strings.Contains(string(body), "AuxiliaryDataNotSupported") {
				fmt.Printf("❌ 미지원: AuxiliaryDataNotSupported\n")
			} else if strings.Contains(string(body), "InvalidArgVal") {
				fmt.Printf("❌ 미지원: InvalidArgVal\n")
			} else {
				fmt.Printf("❌ 실패: %s\n", string(body))
			}
		}

		time.Sleep(300 * time.Millisecond)
	}
}

// 테스트 9: Continuous Focus with integer speed
func testContinuousFocusWithIntSpeed(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	speeds := []int{1, 3, 5, 7, -1, -3, -5, -7}

	for _, speed := range speeds {
		fmt.Printf("\n--- Continuous Focus Speed: %d ---\n", speed)

		req := onvif_imaging.Move{
			VideoSourceToken: token,
			Focus: xsd_onvif.FocusMove{
				Continuous: xsd_onvif.ContinuousFocus{
					Speed: xsd.Float(float64(speed)),
				},
			},
		}

		resp, err := dev.CallMethod(req)
		if err != nil {
			fmt.Printf("❌ 에러: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("응답 코드: %d\n", resp.StatusCode)

		if resp.StatusCode == 200 {
			fmt.Printf("✅ Continuous Focus 성공! (Speed: %d)\n", speed)

			// Stop after brief movement
			time.Sleep(200 * time.Millisecond)
			stopReq := onvif_imaging.Stop{VideoSourceToken: token}
			stopResp, _ := dev.CallMethod(stopReq)
			if stopResp != nil {
				stopResp.Body.Close()
			}
		} else {
			// 에러 메시지 핵심 추출
			if strings.Contains(string(body), "Not support Absolute") {
				fmt.Printf("❌ 실패: Not support Absolute\n")
			} else {
				fmt.Printf("❌ 실패: %s\n", string(body))
			}
		}

		time.Sleep(300 * time.Millisecond)
	}
}

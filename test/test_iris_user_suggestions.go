package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"time"

	"github.com/use-go/onvif"
	onvif_imaging "github.com/use-go/onvif/Imaging"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	onvif_ptz "github.com/use-go/onvif/ptz"
	"github.com/use-go/onvif/xsd"
	xsd_onvif "github.com/use-go/onvif/xsd/onvif"
)

func main() {
	host := "14.51.233.129"
	port := 10082
	username := "admin"
	password := "pluxity123!@#"

	fmt.Printf("=== ONVIF Iris 고급 테스트 (사용자 제안 방법) ===\n\n")

	// Create ONVIF device
	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    fmt.Sprintf("%s:%d", host, port),
		Username: username,
		Password: password,
	})
	if err != nil {
		fmt.Printf("❌ ONVIF 장치 생성 실패: %v\n", err)
		return
	}

	// Get device information
	getInfoReq := device.GetDeviceInformation{}
	_, err = dev.CallMethod(getInfoReq)
	if err != nil {
		fmt.Printf("❌ 장치 정보 조회 실패: %v\n", err)
		return
	}

	// Get media profiles
	getProfilesReq := media.GetProfiles{}
	profilesResp, err := dev.CallMethod(getProfilesReq)
	if err != nil {
		fmt.Printf("❌ 프로필 조회 실패: %v\n", err)
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

	xml.Unmarshal(body, &envelope)
	profile := envelope.Body.GetProfilesResponse.Profiles[0]
	videoSourceToken := xsd_onvif.ReferenceToken(profile.VideoSourceConfiguration.SourceToken)
	profileToken := xsd_onvif.ReferenceToken(profile.Token)

	fmt.Printf("VideoSourceToken: %s\n", videoSourceToken)
	fmt.Printf("ProfileToken: %s\n\n", profileToken)

	// ========================================
	// 테스트 1: Exposure Mode를 MANUAL로만 변경 (Iris 값 제외)
	// ========================================
	fmt.Println("=== 테스트 1: Exposure Mode를 MANUAL로만 변경 (단계별 접근) ===")
	testManualModeOnly(dev, videoSourceToken)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 2: MANUAL 모드 + ExposureTime/Gain/Iris 모두 명시
	// ========================================
	fmt.Println("\n=== 테스트 2: MANUAL 모드 + ExposureTime/Gain/Iris 모두 지정 ===")
	testManualWithAllParams(dev, videoSourceToken)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 3: Imaging Move - Continuous 방식 (Focus로 테스트)
	// ========================================
	fmt.Println("\n=== 테스트 3: Imaging Move - Continuous 방식 (Focus용) ===")
	testImagingMoveContinuous(dev, videoSourceToken)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 4: PTZ GetConfigurationOptions - Auxiliary/Extension 확인
	// ========================================
	// fmt.Println("\n=== 테스트 4: PTZ GetConfigurationOptions ===")
	// testPTZConfigOptions(dev)
	// time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 5-7: PTZ SendAuxiliaryCommand
	// ========================================
	fmt.Println("\n=== 테스트 5: PTZ SendAuxiliaryCommand - IrisOpen ===")
	testAuxiliaryCommand(dev, profileToken, "IrisOpen")
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 테스트 6: PTZ SendAuxiliaryCommand - IrisClose ===")
	testAuxiliaryCommand(dev, profileToken, "IrisClose")
	time.Sleep(2 * time.Second)

	fmt.Println("\n=== 테스트 7: PTZ SendAuxiliaryCommand - Iris Auto ===")
	testAuxiliaryCommand(dev, profileToken, "IrisAuto")
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 8: WDR/BLC를 명시적으로 OFF로 설정 후 Iris 변경
	// ========================================
	fmt.Println("\n=== 테스트 8: WDR/BLC OFF + MANUAL Iris ===")
	testWithWDRBLCOff(dev, videoSourceToken)

	fmt.Println("\n=== 테스트 완료 ===")
}

func testManualModeOnly(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Exposure: xsd_onvif.Exposure20{
				Mode:            xsd_onvif.ExposureMode("MANUAL"),
				MinExposureTime: float64(33),
				MaxExposureTime: float64(33333),
				MinIris:         float64(-22),
				MaxIris:         float64(0),
			},
		},
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ MANUAL 모드 전환 성공!")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 실패 (코드 %d): %s\n", resp.StatusCode, string(body))
	}
}

func testManualWithAllParams(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Exposure: xsd_onvif.Exposure20{
				Mode:         xsd_onvif.ExposureMode("MANUAL"),
				ExposureTime: float64(10000),
				Gain:         float64(50),
				Iris:         float64(-10),
			},
		},
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ MANUAL + 전체 파라미터 설정 성공!")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 실패 (코드 %d): %s\n", resp.StatusCode, string(body))
	}
}

func testImagingMoveContinuous(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.Move{
		VideoSourceToken: token,
		Focus: xsd_onvif.FocusMove{
			Continuous: xsd_onvif.ContinuousFocus{
				Speed: xsd.Float(0.5),
			},
		},
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ Imaging Move (Continuous) 성공!")

		// Stop after 1 second
		time.Sleep(1 * time.Second)
		stopReq := onvif_imaging.Stop{
			VideoSourceToken: token,
		}
		stopResp, _ := dev.CallMethod(stopReq)
		if stopResp != nil {
			stopResp.Body.Close()
			fmt.Println("✅ Imaging Stop 호출 완료")
		}
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 실패 (코드 %d): %s\n", resp.StatusCode, string(body))
	}
}

func testPTZConfigOptions(dev *onvif.Device) {
	// TODO: Fix field name - GetConfigurationOptions 구조체 확인 필요
	fmt.Println("⚠️  테스트 건너뜀: GetConfigurationOptions 필드 구조 확인 필요")
	return
}

func testAuxiliaryCommand(dev *onvif.Device, token xsd_onvif.ReferenceToken, command string) {
	req := onvif_ptz.SendAuxiliaryCommand{
		ProfileToken:  token,
		AuxiliaryData: xsd_onvif.AuxiliaryData(command),
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ %s 명령 성공!\n", command)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 실패 (코드 %d): %s\n", resp.StatusCode, string(body))
	}
}

func testWithWDRBLCOff(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			BacklightCompensation: xsd_onvif.BacklightCompensation20{
				Mode: xsd_onvif.BacklightCompensationMode("OFF"),
			},
			WideDynamicRange: xsd_onvif.WideDynamicRange20{
				Mode: xsd_onvif.WideDynamicMode("OFF"),
			},
			Exposure: xsd_onvif.Exposure20{
				Mode:         xsd_onvif.ExposureMode("MANUAL"),
				ExposureTime: float64(10000),
				Gain:         float64(50),
				Iris:         float64(-10),
			},
		},
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("✅ WDR/BLC OFF + MANUAL Iris 설정 성공!")
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 실패 (코드 %d): %s\n", resp.StatusCode, string(body))
	}
}

func containsAny(text string, keywords []string) bool {
	for _, keyword := range keywords {
		if contains(text, keyword) {
			return true
		}
	}
	return false
}

func contains(text, substr string) bool {
	return len(text) >= len(substr) && findSubstring(text, substr)
}

func findSubstring(text, substr string) bool {
	for i := 0; i <= len(text)-len(substr); i++ {
		if text[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func printRelevantLines(text string, keywords []string) {
	lines := splitLines(text)
	for _, line := range lines {
		for _, keyword := range keywords {
			if contains(line, keyword) {
				fmt.Println(line)
				break
			}
		}
	}
}

func splitLines(text string) []string {
	var lines []string
	var current string
	for _, char := range text {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

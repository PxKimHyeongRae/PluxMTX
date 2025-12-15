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
	xsd_onvif "github.com/use-go/onvif/xsd/onvif"
)

func main() {
	host := "14.51.233.129"
	port := 10082
	username := "admin"
	password := "pluxity123!@#"

	fmt.Println("=== ONVIF Imaging 기본 설정 테스트 ===")
	fmt.Println("=== (밝기, 채도, 명암비, 선명도 등) ===\n")

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

	fmt.Printf("✅ VideoSource Token: %s\n\n", videoSourceToken)

	// ========================================
	// 테스트 1: GetOptions - 지원되는 설정 범위 확인
	// ========================================
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("=== 테스트 1: GetOptions - 설정 범위 확인 ===")
	testGetOptions(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 2: GetImagingSettings - 현재 설정 조회
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 2: GetImagingSettings - 현재 설정 조회 ===")
	currentSettings := testGetImagingSettings(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 3: SetImagingSettings - Brightness만 변경
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 3: SetImagingSettings - Brightness (밝기) ===")
	testSetBrightness(dev, videoSourceToken, 60.0)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 4: SetImagingSettings - ColorSaturation만 변경
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 4: SetImagingSettings - ColorSaturation (채도) ===")
	testSetColorSaturation(dev, videoSourceToken, 60.0)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 5: SetImagingSettings - Contrast만 변경
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 5: SetImagingSettings - Contrast (명암비) ===")
	testSetContrast(dev, videoSourceToken, 60.0)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 6: SetImagingSettings - Sharpness만 변경
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 6: SetImagingSettings - Sharpness (선명도) ===")
	testSetSharpness(dev, videoSourceToken, 60.0)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 7: SetImagingSettings - 여러 설정 동시 변경
	// ========================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 테스트 7: SetImagingSettings - 여러 설정 동시 변경 ===")
	testSetMultipleSettings(dev, videoSourceToken)
	time.Sleep(1 * time.Second)

	// ========================================
	// 테스트 8: 설정 복원 (원래값으로)
	// ========================================
	if currentSettings != nil {
		fmt.Println("\n" + strings.Repeat("=", 60))
		fmt.Println("=== 테스트 8: 설정 복원 ===")
		testRestoreSettings(dev, videoSourceToken, currentSettings)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("=== 모든 테스트 완료 ===")
}

// 테스트 1: GetOptions
func testGetOptions(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.GetOptions{
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
		fmt.Println("✅ GetOptions 성공!")

		// 주요 설정 범위 파싱
		var optEnvelope struct {
			Body struct {
				GetOptionsResponse struct {
					ImagingOptions struct {
						Brightness struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Brightness"`
						ColorSaturation struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"ColorSaturation"`
						Contrast struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Contrast"`
						Sharpness struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Sharpness"`
					} `xml:"ImagingOptions"`
				} `xml:"GetOptionsResponse"`
			} `xml:"Body"`
		}

		if err := xml.Unmarshal(body, &optEnvelope); err == nil {
			opts := optEnvelope.Body.GetOptionsResponse.ImagingOptions
			fmt.Printf("\n📊 지원되는 설정 범위:\n")
			fmt.Printf("   Brightness (밝기):     %.0f ~ %.0f\n", opts.Brightness.Min, opts.Brightness.Max)
			fmt.Printf("   ColorSaturation (채도): %.0f ~ %.0f\n", opts.ColorSaturation.Min, opts.ColorSaturation.Max)
			fmt.Printf("   Contrast (명암비):     %.0f ~ %.0f\n", opts.Contrast.Min, opts.Contrast.Max)
			fmt.Printf("   Sharpness (선명도):    %.0f ~ %.0f\n", opts.Sharpness.Min, opts.Sharpness.Max)
		}
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
}

// 테스트 2: GetImagingSettings
func testGetImagingSettings(dev *onvif.Device, token xsd_onvif.ReferenceToken) *xsd_onvif.ImagingSettings20 {
	req := onvif_imaging.GetImagingSettings{
		VideoSourceToken: token,
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ 에러: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("응답 코드: %d\n", resp.StatusCode)

	if resp.StatusCode == 200 {
		fmt.Println("✅ GetImagingSettings 성공!")

		var settingsEnvelope struct {
			Body struct {
				GetImagingSettingsResponse struct {
					ImagingSettings xsd_onvif.ImagingSettings20
				} `xml:"GetImagingSettingsResponse"`
			} `xml:"Body"`
		}

		if err := xml.Unmarshal(body, &settingsEnvelope); err == nil {
			settings := settingsEnvelope.Body.GetImagingSettingsResponse.ImagingSettings
			fmt.Printf("\n📊 현재 설정값:\n")
			fmt.Printf("   Brightness (밝기):     %.1f\n", settings.Brightness)
			fmt.Printf("   ColorSaturation (채도): %.1f\n", settings.ColorSaturation)
			fmt.Printf("   Contrast (명암비):     %.1f\n", settings.Contrast)
			fmt.Printf("   Sharpness (선명도):    %.1f\n", settings.Sharpness)
			return &settings
		}
	} else {
		fmt.Printf("❌ 실패: %s\n", string(body))
	}
	return nil
}

// 테스트 3: Brightness 변경
func testSetBrightness(dev *onvif.Device, token xsd_onvif.ReferenceToken, value float64) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Brightness: value,
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
		fmt.Printf("✅ Brightness = %.1f 설정 성공!\n", value)
	} else {
		// 에러 메시지 추출
		if strings.Contains(string(body), "Invalid BLC") {
			fmt.Printf("❌ 실패: Invalid BLC\n")
		} else {
			fmt.Printf("❌ 실패: %s\n", extractErrorMessage(string(body)))
		}
	}
}

// 테스트 4: ColorSaturation 변경
func testSetColorSaturation(dev *onvif.Device, token xsd_onvif.ReferenceToken, value float64) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			ColorSaturation: value,
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
		fmt.Printf("✅ ColorSaturation = %.1f 설정 성공!\n", value)
	} else {
		if strings.Contains(string(body), "Invalid BLC") {
			fmt.Printf("❌ 실패: Invalid BLC\n")
		} else {
			fmt.Printf("❌ 실패: %s\n", extractErrorMessage(string(body)))
		}
	}
}

// 테스트 5: Contrast 변경
func testSetContrast(dev *onvif.Device, token xsd_onvif.ReferenceToken, value float64) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Contrast: value,
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
		fmt.Printf("✅ Contrast = %.1f 설정 성공!\n", value)
	} else {
		if strings.Contains(string(body), "Invalid BLC") {
			fmt.Printf("❌ 실패: Invalid BLC\n")
		} else {
			fmt.Printf("❌ 실패: %s\n", extractErrorMessage(string(body)))
		}
	}
}

// 테스트 6: Sharpness 변경
func testSetSharpness(dev *onvif.Device, token xsd_onvif.ReferenceToken, value float64) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Sharpness: value,
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
		fmt.Printf("✅ Sharpness = %.1f 설정 성공!\n", value)
	} else {
		if strings.Contains(string(body), "Invalid BLC") {
			fmt.Printf("❌ 실패: Invalid BLC\n")
		} else {
			fmt.Printf("❌ 실패: %s\n", extractErrorMessage(string(body)))
		}
	}
}

// 테스트 7: 여러 설정 동시 변경
func testSetMultipleSettings(dev *onvif.Device, token xsd_onvif.ReferenceToken) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Brightness:      55.0,
			ColorSaturation: 55.0,
			Contrast:        55.0,
			Sharpness:       55.0,
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
		fmt.Println("✅ 여러 설정 동시 변경 성공!")
		fmt.Println("   Brightness=55, ColorSaturation=55, Contrast=55, Sharpness=55")
	} else {
		if strings.Contains(string(body), "Invalid BLC") {
			fmt.Printf("❌ 실패: Invalid BLC\n")
		} else {
			fmt.Printf("❌ 실패: %s\n", extractErrorMessage(string(body)))
		}
	}
}

// 테스트 8: 설정 복원
func testRestoreSettings(dev *onvif.Device, token xsd_onvif.ReferenceToken, original *xsd_onvif.ImagingSettings20) {
	req := onvif_imaging.SetImagingSettings{
		VideoSourceToken: token,
		ImagingSettings: xsd_onvif.ImagingSettings20{
			Brightness:      original.Brightness,
			ColorSaturation: original.ColorSaturation,
			Contrast:        original.Contrast,
			Sharpness:       original.Sharpness,
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
		fmt.Println("✅ 원래 설정으로 복원 성공!")
		fmt.Printf("   Brightness=%.1f, ColorSaturation=%.1f, Contrast=%.1f, Sharpness=%.1f\n",
			original.Brightness, original.ColorSaturation, original.Contrast, original.Sharpness)
	} else {
		if strings.Contains(string(body), "Invalid BLC") {
			fmt.Printf("❌ 실패: Invalid BLC\n")
		} else {
			fmt.Printf("❌ 실패: %s\n", extractErrorMessage(string(body)))
		}
	}
}

// 에러 메시지 추출
func extractErrorMessage(xmlBody string) string {
	if strings.Contains(xmlBody, "<env:Text>") {
		start := strings.Index(xmlBody, "<env:Text>")
		end := strings.Index(xmlBody, "</env:Text>")
		if start != -1 && end != -1 && end > start {
			return xmlBody[start+10 : end]
		}
	}
	if len(xmlBody) > 200 {
		return xmlBody[:200] + "..."
	}
	return xmlBody
}

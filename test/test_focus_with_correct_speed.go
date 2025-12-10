package main

import (
	"fmt"
	"io"
	"time"

	"github.com/use-go/onvif"
	onvif_imaging "github.com/use-go/onvif/Imaging"
	onvif_media "github.com/use-go/onvif/media"
	"github.com/use-go/onvif/xsd"
	xsd_onvif "github.com/use-go/onvif/xsd/onvif"
)

func main() {
	host := "14.51.233.129"
	port := 10081
	username := "admin"
	password := "pluxity123!@#"

	fmt.Printf("=== ONVIF Focus 올바른 Speed로 테스트 ===\n\n")

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

	// Get VideoSourceToken
	getProfilesReq := onvif_media.GetProfiles{}
	profilesResp, err := dev.CallMethod(getProfilesReq)
	if err != nil {
		fmt.Printf("❌ GetProfiles 실패: %v\n", err)
		return
	}

	// VideoSourceToken 추출 (간략화)
	videoSourceToken := xsd_onvif.ReferenceToken("VideoSource_1")
	fmt.Printf("VideoSource Token: %s\n\n", videoSourceToken)
	profilesResp.Body.Close()

	// ========================================
	// 테스트 1: Speed = 1.0 (정규화된 값)
	// ========================================
	fmt.Println("=== 테스트 1: Move Focus Far with Speed 1.0 ===")
	testFocusMove(dev, videoSourceToken, 1.0)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 2: Speed = 5 (정수 값, GetMoveOptions 범위 내)
	// ========================================
	fmt.Println("\n=== 테스트 2: Move Focus Far with Speed 5 ===")
	testFocusMove(dev, videoSourceToken, 5.0)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 3: Speed = 3 (중간 값)
	// ========================================
	fmt.Println("\n=== 테스트 3: Move Focus Far with Speed 3 ===")
	testFocusMove(dev, videoSourceToken, 3.0)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 4: Speed = -3 (근거리 포커스)
	// ========================================
	fmt.Println("\n=== 테스트 4: Move Focus Near with Speed -3 ===")
	testFocusMove(dev, videoSourceToken, -3.0)
	time.Sleep(2 * time.Second)

	// ========================================
	// 테스트 5: Stop
	// ========================================
	fmt.Println("\n=== 테스트 5: Stop Focus ===")
	stopReq := onvif_imaging.Stop{
		VideoSourceToken: videoSourceToken,
	}

	stopResp, err := dev.CallMethod(stopReq)
	if err != nil {
		fmt.Printf("❌ Stop 에러: %v\n", err)
	} else {
		defer stopResp.Body.Close()
		if stopResp.StatusCode == 200 {
			fmt.Println("✅ Stop 성공!")
		} else {
			body, _ := io.ReadAll(stopResp.Body)
			fmt.Printf("❌ Stop 실패 (코드 %d): %s\n", stopResp.StatusCode, string(body))
		}
	}

	fmt.Println("\n=== 테스트 완료 ===")
}

func testFocusMove(dev *onvif.Device, token xsd_onvif.ReferenceToken, speed float64) {
	req := onvif_imaging.Move{
		VideoSourceToken: token,
		Focus: xsd_onvif.FocusMove{
			Continuous: xsd_onvif.ContinuousFocus{
				Speed: xsd.Float(speed),
			},
		},
	}

	resp, err := dev.CallMethod(req)
	if err != nil {
		fmt.Printf("❌ Move 에러: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Printf("✅ Focus Move 성공! (Speed: %.1f)\n", speed)
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ Focus Move 실패 (코드 %d): %s\n", resp.StatusCode, string(body))
	}
}

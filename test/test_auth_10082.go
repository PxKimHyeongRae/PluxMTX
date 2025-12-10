package main

import (
	"fmt"
	"io"

	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
)

func main() {
	host := "14.51.233.129"
	port := 10082
	username := "admin"
	password := "pluxity123!@#"

	fmt.Printf("=== ONVIF Digest 인증 테스트 (포트 %d) ===\n\n", port)

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

	fmt.Println("✅ ONVIF 장치 생성 성공")

	// Test 1: GetDeviceInformation
	fmt.Println("\n=== 테스트 1: GetDeviceInformation ===")
	getInfoReq := device.GetDeviceInformation{}
	resp, err := dev.CallMethod(getInfoReq)
	if err != nil {
		fmt.Printf("❌ 장치 정보 조회 실패: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("✅ Digest 인증 성공! (Status 200)")
		fmt.Printf("\n응답 내용:\n%s\n", string(body))
	} else if resp.StatusCode == 401 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ 인증 실패 (Status 401)\n%s\n", string(body))
	} else {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("⚠️  예상치 못한 응답 (Status %d)\n%s\n", resp.StatusCode, string(body))
	}
}

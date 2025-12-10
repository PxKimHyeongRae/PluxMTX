package main

import (
	"encoding/xml"
	"fmt"
	"io"

	"github.com/use-go/onvif"
	onvif_imaging "github.com/use-go/onvif/Imaging"
	onvif_media "github.com/use-go/onvif/media"
	xsd_onvif "github.com/use-go/onvif/xsd/onvif"
)

func main() {
	host := "14.51.233.129"
	port := 10081
	username := "admin"
	password := "pluxity123!@#"

	fmt.Printf("=== ONVIF Focus GetMoveOptions 테스트 ===\n\n")

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

	fmt.Printf("✅ VideoSource Token: %s\n\n", videoSourceToken)

	// ========================================
	// 테스트: GetMoveOptions - 카메라가 지원하는 Focus 제어 모드 확인
	// ========================================
	fmt.Println("=== GetMoveOptions - Focus 지원 모드 확인 ===")
	getMoveOptionsReq := onvif_imaging.GetMoveOptions{
		VideoSourceToken: videoSourceToken,
	}

	moveOptionsResp, err := dev.CallMethod(getMoveOptionsReq)
	if err != nil {
		fmt.Printf("❌ GetMoveOptions 실패: %v\n", err)
		return
	}

	body, _ = io.ReadAll(moveOptionsResp.Body)
	moveOptionsResp.Body.Close()

	fmt.Printf("응답 상태: %s\n", moveOptionsResp.Status)
	fmt.Printf("\n전체 응답:\n%s\n\n", string(body))

	// Parse the response to extract Focus options
	var moveOptsEnvelope struct {
		Body struct {
			GetMoveOptionsResponse struct {
				MoveOptions struct {
					Absolute struct {
						Position struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Position"`
						Speed struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Speed"`
					} `xml:"Absolute"`
					Relative struct {
						Distance struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Distance"`
						Speed struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Speed"`
					} `xml:"Relative"`
					Continuous struct {
						Speed struct {
							Min float64 `xml:"Min"`
							Max float64 `xml:"Max"`
						} `xml:"Speed"`
					} `xml:"Continuous"`
				} `xml:"MoveOptions"`
			} `xml:"GetMoveOptionsResponse"`
		} `xml:"Body"`
	}

	if err := xml.Unmarshal(body, &moveOptsEnvelope); err != nil {
		fmt.Printf("⚠️  파싱 실패 (구조가 다를 수 있음): %v\n", err)
	} else {
		opts := moveOptsEnvelope.Body.GetMoveOptionsResponse.MoveOptions

		fmt.Println("=== 파싱된 Focus 제어 옵션 ===")

		// Check Absolute
		if opts.Absolute.Position.Max != 0 || opts.Absolute.Position.Min != 0 {
			fmt.Printf("✅ Absolute Focus 지원:\n")
			fmt.Printf("   Position: %.2f ~ %.2f\n", opts.Absolute.Position.Min, opts.Absolute.Position.Max)
			if opts.Absolute.Speed.Max != 0 {
				fmt.Printf("   Speed: %.2f ~ %.2f\n", opts.Absolute.Speed.Min, opts.Absolute.Speed.Max)
			}
		} else {
			fmt.Println("❌ Absolute Focus 미지원")
		}

		// Check Relative
		if opts.Relative.Distance.Max != 0 || opts.Relative.Distance.Min != 0 {
			fmt.Printf("✅ Relative Focus 지원:\n")
			fmt.Printf("   Distance: %.2f ~ %.2f\n", opts.Relative.Distance.Min, opts.Relative.Distance.Max)
			if opts.Relative.Speed.Max != 0 {
				fmt.Printf("   Speed: %.2f ~ %.2f\n", opts.Relative.Speed.Min, opts.Relative.Speed.Max)
			}
		} else {
			fmt.Println("❌ Relative Focus 미지원")
		}

		// Check Continuous
		if opts.Continuous.Speed.Max != 0 || opts.Continuous.Speed.Min != 0 {
			fmt.Printf("✅ Continuous Focus 지원:\n")
			fmt.Printf("   Speed: %.2f ~ %.2f\n", opts.Continuous.Speed.Min, opts.Continuous.Speed.Max)
		} else {
			fmt.Println("❌ Continuous Focus 미지원")
		}
	}

	fmt.Println("\n=== 테스트 완료 ===")
}

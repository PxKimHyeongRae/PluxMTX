package main

import (
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	cameraIP   = "14.51.233.129"
	cameraPort = "10081"
	username   = "admin"
	password   = "p2scctv!@"
)

var baseURL = fmt.Sprintf("http://%s:%s/onvif/device_service", cameraIP, cameraPort)
var imagingURL = fmt.Sprintf("http://%s:%s/onvif/imaging_service", cameraIP, cameraPort)
var ptzURL = fmt.Sprintf("http://%s:%s/onvif/ptz_service", cameraIP, cameraPort)

func digestAuth(realm, nonce, uri, method string) string {
	ha1 := fmt.Sprintf("%x", sha256.Sum256([]byte(username+":"+realm+":"+password)))
	ha2 := fmt.Sprintf("%x", sha256.Sum256([]byte(method+":"+uri)))
	response := fmt.Sprintf("%x", sha256.Sum256([]byte(ha1+":"+nonce+":"+ha2)))
	return fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", response="%s"`,
		username, realm, nonce, uri, response)
}

func sendRequest(url, soapAction, body string) (int, string) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
	if soapAction != "" {
		req.Header.Set("SOAPAction", soapAction)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Sprintf("Error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		authHeader := resp.Header.Get("WWW-Authenticate")
		if strings.Contains(authHeader, "Digest") {
			var realm, nonce string
			for _, part := range strings.Split(authHeader, ",") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "realm=") {
					realm = strings.Trim(strings.TrimPrefix(part, "realm="), `"`)
				} else if strings.HasPrefix(part, "nonce=") {
					nonce = strings.Trim(strings.TrimPrefix(part, "nonce="), `"`)
				}
			}

			req, _ = http.NewRequest("POST", url, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/soap+xml; charset=utf-8")
			req.Header.Set("Authorization", digestAuth(realm, nonce, req.URL.RequestURI(), "POST"))
			if soapAction != "" {
				req.Header.Set("SOAPAction", soapAction)
			}

			resp, err = client.Do(req)
			if err != nil {
				return 0, fmt.Sprintf("Error: %v", err)
			}
			defer resp.Body.Close()
		}
	}

	bodyBytes, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(bodyBytes)
}

func getVideoSourceToken() string {
	body := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
		<GetProfiles xmlns="http://www.onvif.org/ver10/media/wsdl"/>
	</s:Body>
</s:Envelope>`

	mediaURL := fmt.Sprintf("http://%s:%s/onvif/media_service", cameraIP, cameraPort)
	_, respBody := sendRequest(mediaURL, "", body)

	type Profile struct {
		Token                    string `xml:"token,attr"`
		VideoSourceConfiguration struct {
			SourceToken string
		}
	}
	var envelope struct {
		Body struct {
			GetProfilesResponse struct {
				Profiles []Profile
			}
		}
	}

	xml.Unmarshal([]byte(respBody), &envelope)
	if len(envelope.Body.GetProfilesResponse.Profiles) > 0 {
		return envelope.Body.GetProfilesResponse.Profiles[0].VideoSourceConfiguration.SourceToken
	}
	return ""
}

func getPTZToken() string {
	body := `<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope">
	<s:Body xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
		<GetProfiles xmlns="http://www.onvif.org/ver10/media/wsdl"/>
	</s:Body>
</s:Envelope>`

	mediaURL := fmt.Sprintf("http://%s:%s/onvif/media_service", cameraIP, cameraPort)
	_, respBody := sendRequest(mediaURL, "", body)

	type Profile struct {
		Token string `xml:"token,attr"`
	}
	var envelope struct {
		Body struct {
			GetProfilesResponse struct {
				Profiles []Profile
			}
		}
	}

	xml.Unmarshal([]byte(respBody), &envelope)
	if len(envelope.Body.GetProfilesResponse.Profiles) > 0 {
		return envelope.Body.GetProfilesResponse.Profiles[0].Token
	}
	return ""
}

func main() {
	fmt.Println("=== ONVIF Iris 고급 테스트 (사용자 제안 방법) ===\n")

	videoSourceToken := getVideoSourceToken()
	ptzToken := getPTZToken()
	fmt.Printf("VideoSourceToken: %s\n", videoSourceToken)
	fmt.Printf("PTZToken: %s\n\n", ptzToken)

	// ========================================
	// 테스트 1: Exposure Mode를 MANUAL로만 변경 (Iris 값은 나중에)
	// ========================================
	fmt.Println("=== 테스트 1: Exposure Mode를 MANUAL로만 변경 (단계별 접근) ===")
	body1 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:timg="http://www.onvif.org/ver20/imaging/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<timg:SetImagingSettings>
			<timg:VideoSourceToken>%s</timg:VideoSourceToken>
			<timg:ImagingSettings>
				<tt:Exposure>
					<tt:Mode>MANUAL</tt:Mode>
					<tt:MinExposureTime>33</tt:MinExposureTime>
					<tt:MaxExposureTime>33333</tt:MaxExposureTime>
					<tt:MinIris>-22</tt:MinIris>
					<tt:MaxIris>0</tt:MaxIris>
				</tt:Exposure>
			</timg:ImagingSettings>
		</timg:SetImagingSettings>
	</s:Body>
</s:Envelope>`, videoSourceToken)

	code1, resp1 := sendRequest(imagingURL, "", body1)
	fmt.Printf("응답 코드: %d\n", code1)
	if code1 == 200 {
		fmt.Println("✅ MANUAL 모드 전환 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp1)
	}
	fmt.Println()

	// ========================================
	// 테스트 2: Imaging Move - Continuous 방식 (Absolute가 아닌 Speed 기반)
	// ========================================
	fmt.Println("=== 테스트 2: Imaging Move - Continuous 방식 (Speed 기반) ===")
	body2 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:timg="http://www.onvif.org/ver20/imaging/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<timg:Move>
			<timg:VideoSourceToken>%s</timg:VideoSourceToken>
			<timg:Focus>
				<tt:Continuous>
					<tt:Speed>0.5</tt:Speed>
				</tt:Continuous>
			</timg:Focus>
		</timg:Move>
	</s:Body>
</s:Envelope>`, videoSourceToken)

	code2, resp2 := sendRequest(imagingURL, "", body2)
	fmt.Printf("응답 코드: %d\n", code2)
	if code2 == 200 {
		fmt.Println("✅ Imaging Move (Continuous) 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp2)
	}
	fmt.Println()

	// ========================================
	// 테스트 3: Imaging Stop (Move 후 정지)
	// ========================================
	fmt.Println("=== 테스트 3: Imaging Stop ===")
	body3 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:timg="http://www.onvif.org/ver20/imaging/wsdl">
	<s:Body>
		<timg:Stop>
			<timg:VideoSourceToken>%s</timg:VideoSourceToken>
		</timg:Stop>
	</s:Body>
</s:Envelope>`, videoSourceToken)

	code3, resp3 := sendRequest(imagingURL, "", body3)
	fmt.Printf("응답 코드: %d\n", code3)
	if code3 == 200 {
		fmt.Println("✅ Imaging Stop 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp3)
	}
	fmt.Println()

	// ========================================
	// 테스트 4: PTZ GetConfigurationOptions - Auxiliary 지원 확인
	// ========================================
	fmt.Println("=== 테스트 4: PTZ GetConfigurationOptions ===")
	body4 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl">
	<s:Body>
		<tptz:GetConfigurationOptions>
			<tptz:ConfigurationToken>PTZConfiguration_1</tptz:ConfigurationToken>
		</tptz:GetConfigurationOptions>
	</s:Body>
</s:Envelope>`)

	code4, resp4 := sendRequest(ptzURL, "", body4)
	fmt.Printf("응답 코드: %d\n", code4)
	if code4 == 200 {
		fmt.Println("✅ PTZ Configuration Options 조회 성공")
		// PTZSpaces나 Extension 확인
		if strings.Contains(resp4, "Iris") || strings.Contains(resp4, "Auxiliary") {
			fmt.Println("🔍 Iris 또는 Auxiliary 관련 항목 발견!")
		}
		fmt.Println("\n관련 항목 검색:")
		lines := strings.Split(resp4, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Auxiliary") || strings.Contains(line, "Iris") ||
				strings.Contains(line, "Extension") || strings.Contains(line, "Space") {
				fmt.Println(line)
			}
		}
	} else {
		fmt.Printf("❌ 실패: %s\n", resp4)
	}
	fmt.Println()

	// ========================================
	// 테스트 5: PTZ SendAuxiliaryCommand - IrisOpen
	// ========================================
	fmt.Println("=== 테스트 5: PTZ SendAuxiliaryCommand - IrisOpen ===")
	body5 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<tptz:SendAuxiliaryCommand>
			<tptz:ProfileToken>%s</tptz:ProfileToken>
			<tptz:AuxiliaryData>IrisOpen</tptz:AuxiliaryData>
		</tptz:SendAuxiliaryCommand>
	</s:Body>
</s:Envelope>`, ptzToken)

	code5, resp5 := sendRequest(ptzURL, "", body5)
	fmt.Printf("응답 코드: %d\n", code5)
	if code5 == 200 {
		fmt.Println("✅ IrisOpen 명령 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp5)
	}
	fmt.Println()

	// ========================================
	// 테스트 6: PTZ SendAuxiliaryCommand - IrisClose
	// ========================================
	fmt.Println("=== 테스트 6: PTZ SendAuxiliaryCommand - IrisClose ===")
	body6 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<tptz:SendAuxiliaryCommand>
			<tptz:ProfileToken>%s</tptz:ProfileToken>
			<tptz:AuxiliaryData>IrisClose</tptz:AuxiliaryData>
		</tptz:SendAuxiliaryCommand>
	</s:Body>
</s:Envelope>`, ptzToken)

	code6, resp6 := sendRequest(ptzURL, "", body6)
	fmt.Printf("응답 코드: %d\n", code6)
	if code6 == 200 {
		fmt.Println("✅ IrisClose 명령 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp6)
	}
	fmt.Println()

	// ========================================
	// 테스트 7: PTZ SendAuxiliaryCommand - Iris Auto
	// ========================================
	fmt.Println("=== 테스트 7: PTZ SendAuxiliaryCommand - Iris Auto ===")
	body7 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl">
	<s:Body>
		<tptz:SendAuxiliaryCommand>
			<tptz:ProfileToken>%s</tptz:ProfileToken>
			<tptz:AuxiliaryData>Iris Auto</tptz:AuxiliaryData>
		</tptz:SendAuxiliaryCommand>
	</s:Body>
</s:Envelope>`, ptzToken)

	code7, resp7 := sendRequest(ptzURL, "", body7)
	fmt.Printf("응답 코드: %d\n", code7)
	if code7 == 200 {
		fmt.Println("✅ Iris Auto 명령 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp7)
	}
	fmt.Println()

	// ========================================
	// 테스트 8: MANUAL 모드에서 ExposureTime과 Iris를 명시적으로 설정
	// ========================================
	fmt.Println("=== 테스트 8: MANUAL 모드 + ExposureTime/Gain/Iris 모두 지정 ===")
	body8 := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<s:Envelope xmlns:s="http://www.w3.org/2003/05/soap-envelope" xmlns:timg="http://www.onvif.org/ver20/imaging/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema">
	<s:Body>
		<timg:SetImagingSettings>
			<timg:VideoSourceToken>%s</timg:VideoSourceToken>
			<timg:ImagingSettings>
				<tt:Exposure>
					<tt:Mode>MANUAL</tt:Mode>
					<tt:ExposureTime>10000</tt:ExposureTime>
					<tt:Gain>50</tt:Gain>
					<tt:Iris>-10</tt:Iris>
				</tt:Exposure>
			</timg:ImagingSettings>
		</timg:SetImagingSettings>
	</s:Body>
</s:Envelope>`, videoSourceToken)

	code8, resp8 := sendRequest(imagingURL, "", body8)
	fmt.Printf("응답 코드: %d\n", code8)
	if code8 == 200 {
		fmt.Println("✅ MANUAL 모드 + Iris 설정 성공!")
	} else {
		fmt.Printf("❌ 실패: %s\n", resp8)
	}
	fmt.Println()

	fmt.Println("=== 테스트 완료 ===")
}

package wecom

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const baseURL = "https://qyapi.weixin.qq.com/cgi-bin"

// AccessTokenResponse gettoken响应
type AccessTokenResponse struct {
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// UserInfoResponse getuserinfo响应
type UserInfoResponse struct {
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
	UserID   string `json:"UserId"`
	DeviceID string `json:"DeviceId"`
}

// WeComUser 企微用户详细信息
type WeComUser struct {
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
	UserID     string `json:"userid"`
	Name       string `json:"name"`
	Mobile     string `json:"mobile"`
	Email      string `json:"email"`
	Department []int  `json:"department"`
	Position   string `json:"position"`
	Avatar     string `json:"avatar"`
}

// GetAccessToken 获取企业微信access_token
func GetAccessToken(corpID, secret string) (*AccessTokenResponse, error) {
	u := fmt.Sprintf("%s/gettoken?corpid=%s&corpsecret=%s",
		baseURL, url.QueryEscape(corpID), url.QueryEscape(secret))

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("请求企微gettoken失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取gettoken响应失败: %w", err)
	}

	var result AccessTokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析gettoken响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("企微gettoken错误: errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// GetUserIDByCode 通过OAuth code获取企微UserID
func GetUserIDByCode(accessToken, code string) (*UserInfoResponse, error) {
	u := fmt.Sprintf("%s/user/getuserinfo?access_token=%s&code=%s",
		baseURL, url.QueryEscape(accessToken), url.QueryEscape(code))

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("请求企微getuserinfo失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取getuserinfo响应失败: %w", err)
	}

	var result UserInfoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析getuserinfo响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("企微getuserinfo错误: errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// DepartmentListResponse 部门列表响应
type DepartmentListResponse struct {
	ErrCode    int              `json:"errcode"`
	ErrMsg     string           `json:"errmsg"`
	Department []DepartmentItem `json:"department"`
}

// DepartmentItem 部门信息
type DepartmentItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parentid"`
	Order    int    `json:"order"`
}

// GetDepartmentName 根据部门ID获取部门名称
func GetDepartmentName(accessToken string, deptID int) string {
	u := fmt.Sprintf("%s/department/list?access_token=%s&id=%d",
		baseURL, url.QueryEscape(accessToken), deptID)

	resp, err := http.Get(u)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var result DepartmentListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	if result.ErrCode != 0 || len(result.Department) == 0 {
		return ""
	}

	return result.Department[0].Name
}

// GetUserDetail 获取企微用户详细信息
func GetUserDetail(accessToken, userID string) (*WeComUser, error) {
	u := fmt.Sprintf("%s/user/get?access_token=%s&userid=%s",
		baseURL, url.QueryEscape(accessToken), url.QueryEscape(userID))

	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("请求企微user/get失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取user/get响应失败: %w", err)
	}

	var result WeComUser
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析user/get响应失败: %w", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("企微user/get错误: errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

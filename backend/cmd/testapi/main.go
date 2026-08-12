package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func main() {
	// 测试注册
	regBody := `{"username":"test123","password":"123456","name":"Test"}`
	resp, err := http.Post("http://localhost:8080/api/v1/auth/register", "application/json", bytes.NewBufferString(regBody))
	if err != nil {
		fmt.Println("REG ERROR:", err)
		return
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("REGISTER:", resp.Status, string(body))

	// 测试登录
	loginBody := `{"username":"test123","password":"123456"}`
	resp2, err := http.Post("http://localhost:8080/api/v1/auth/login", "application/json", bytes.NewBufferString(loginBody))
	if err != nil {
		fmt.Println("LOGIN ERROR:", err)
		return
	}
	body2, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	fmt.Println("LOGIN:", resp2.Status, string(body2))
}

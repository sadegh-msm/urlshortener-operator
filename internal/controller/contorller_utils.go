package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func shortenURL(longURL string, expireAt *metav1.Time) (string, error) {
	url := ShortenerServiceURL + "/shorten"

	payload := map[string]string{
		"long_url": longURL,
	}
	if expireAt != nil {
		payload["expire_at"] = expireAt.Time.Format(time.RFC3339)
	} else {
		payload["expire_at"] = ""
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]string
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result["short_url"], nil
}

func getClickCount(shortURL string) (int, error) {
	url := ShortenerServiceURL + "/count/" + shortURL

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result map[string]int
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	return result["click_count"], nil
}

func checkURLValidity(shortURL string) (string, error) {
	url := ShortenerServiceURL + "/valid/" + shortURL

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]bool
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	isValid, exists := result["is_valid"]
	if !exists {
		return "", fmt.Errorf("unexpected response format")
	}

	valid := strconv.FormatBool(isValid)

	return valid, nil
}

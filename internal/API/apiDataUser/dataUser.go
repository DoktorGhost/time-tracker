package apiDataUser

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time-tracker/internal/models"
)

func GetPeopleInfoFromAPI(series, number, urlAPI string) (*models.UserData, error) {
	url := fmt.Sprintf("%s/info?passportSerie=%s&passportNumber=%s",
		urlAPI, series, number)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса к API: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("неправильный статус код API: %d", resp.StatusCode)
	}

	var userInfo models.UserData
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %v", err)
	}
	userInfo.PassportNumber = number
	userInfo.PassportSeries = series

	return &userInfo, nil
}

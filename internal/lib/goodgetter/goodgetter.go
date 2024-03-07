package goodgetter

import (
	"encoding/json"
	"inHouseAd/internal/http-server/handlers/goodsservice/good"
	"io"
	"log/slog"
	"net/http"
)

type Data struct {
	Msg string `json:"msg"`
}

func GetGoodFromAPI(log *slog.Logger, apiURL string, adderGood good.AdderGood) {
	const op = "internal.lib.goodgetter.GetGoodFromAPI"

	log = slog.With(slog.String("op", op))

	var data Data

	for i := 0; i < 3; i++ {
		resp, err := http.Post(apiURL, "", nil)
		if err != nil {
			log.Error("no response from request", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("error: ", err)
			return
		}

		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Error("error: ", err)
			return
		}

		_, _, err = adderGood.AddGood(data.Msg, 1)
		if err != nil {
			log.Error("failed to create category: ", err)

			return
		}
	}
	log.Info("3 goods added")
}

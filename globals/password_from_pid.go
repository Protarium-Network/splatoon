package globals

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"os"
	"strconv"

	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/nex-protocols-go/v2/globals"

	"github.com/PretendoNetwork/nex-go/v2"
)

// PasswordFromPID derives the NEX password as HMAC-SHA256(secret, pid as 8 bytes LE) so the
// token service and this server agree on it without storing credentials.
func PasswordFromPID(pid types.PID) (string, uint32) {
	if len(NEXPasswordSecret) < 32 {
		return "", nex.ResultCodes.RendezVous.InvalidUsername
	}

	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(pid))
	mac := hmac.New(sha256.New, NEXPasswordSecret)
	_, _ = mac.Write(b)

	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), 0
}

// This is the same format as nex-viewer's settings.json
type jsonAccount struct {
	Platform string  `json:"platform"`
	Username string  `json:"username"`
	Pid      float64 `json:"pid"`
	Password string  `json:"password"`
}

type settingsJson struct {
	Accounts []jsonAccount `json:"accounts"`
}

// PasswordFromPIDLocal is an alternative NEX password validator that can be used offline
func PasswordFromPIDLocal(pid types.PID) (string, uint32) {
	file, err := os.ReadFile("settings.json")
	if err != nil {
		Logger.Error(err.Error())
		return "", nex.ResultCodes.RendezVous.InvalidUsername
	}

	var data *settingsJson
	err = json.Unmarshal(file, &data)
	if err != nil {
		Logger.Error(err.Error())
		return "", nex.ResultCodes.RendezVous.InvalidUsername
	}

	for _, account := range data.Accounts {
		if account.Username == strconv.FormatUint(uint64(pid), 10) {
			globals.Logger.Infof("Using local account details for %v", account.Username)
			return account.Password, 0
		}
	}

	return "", nex.ResultCodes.RendezVous.InvalidUsername
}

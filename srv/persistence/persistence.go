package persistence //import "go.iondynamics.net/siteMgr/srv/persistence"

import (
	"encoding/json"

	"go.iondynamics.net/iDhelper/crypto"
	idl "go.iondynamics.net/iDlogger"
	kv "go.iondynamics.net/kvStore"

	"go.iondynamics.net/siteMgr"
)

func ReadUser(user string, pass string) (*siteMgr.User, error) {
	var str string
	err := kv.Read([]byte("user"), []byte(user), &str)
	if err != nil {
		return nil, err
	}

	str = crypto.Decrypt(pass, str)
	usr := siteMgr.NewUser()
	return usr, json.Unmarshal([]byte(str), usr)
}

func UpsertUser(u *siteMgr.User) error {
	byt, err := json.Marshal(u)
	if err != nil {
		return err
	}

	idl.Debug(u)
	idl.Debug(string(byt))

	str := crypto.Encrypt(u.Password, string(byt))
	return kv.Upsert([]byte("user"), []byte(u.Name), str)
}

func DeleteUser(user string) error {
	return kv.Delete([]byte("user"), []byte(user))
}

func UserExists(user string) (bool, error) {
	return kv.Exists([]byte("user"), []byte(user))
}

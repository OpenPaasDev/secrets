package secrets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/gopenpgp/v2/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var secretPwd = []byte("LongSecret")
var johnDoeKey *crypto.Key
var janeDoeKey *crypto.Key
var steveDoeKey *crypto.Key
var johnArmored string
var janeArmored string
var steveArmored string

func init() {
	theKey, err := helper.GenerateKey("John Doe", "john@doe.com", secretPwd, "rsa", 4096)
	if err != nil {
		panic(err)
	}
	johnArmored = theKey
	johnDoeKey, err = crypto.NewKeyFromArmored(johnArmored)
	if err != nil {
		panic(err)
	}

	janeKeyStr, err := helper.GenerateKey("Jane Doe", "jane@doe.com", secretPwd, "rsa", 4096)

	janeArmored = janeKeyStr
	if err != nil {
		panic(err)
	}
	janeDoeKey, err = crypto.NewKeyFromArmored(janeKeyStr)
	if err != nil {
		panic(err)
	}

	steve, err := helper.GenerateKey("Steve Doe", "Steve@doe.com", secretPwd, "rsa", 4096)

	steveArmored = steve
	if err != nil {
		panic(err)
	}
	steveDoeKey, err = crypto.NewKeyFromArmored(steveArmored)
	if err != nil {
		panic(err)
	}
}

func Test_Encrypt_Decrypt_To_Multiple_Keys(t *testing.T) {
	err := os.Setenv("OPENPAAS_PASSPHRASE", "LongSecret")
	require.NoError(t, err)
	envName := RandStringBytes()
	err = os.MkdirAll(filepath.Join("testdata", envName, "secrets"), 0750)
	require.NoError(t, err)
	err = writePubKey("testdata", envName, johnDoeKey)
	require.NoError(t, err)
	err = writePubKey("testdata", envName, janeDoeKey)
	require.NoError(t, err)

	err = WriteSecret("testdata", envName, "FOO", "foobar")
	require.NoError(t, err)
	err = WriteSecret("testdata", envName, "BAR", "bazqux")
	require.NoError(t, err)

	secrets, err := getAllSecretsPrivate(johnArmored, "testdata", envName)
	require.NoError(t, err)
	assert.Len(t, secrets, 2)

	secrets2, err := getAllSecretsPrivate(janeArmored, "testdata", envName)
	require.NoError(t, err)

	_, err = getAllSecretsPrivate(steveArmored, "testdata", envName)
	require.Error(t, err)

	assert.Equal(t, secrets, secrets2)

	assert.Equal(t, Secret{Name: "BAR", Value: "bazqux"}, *secrets[0])

	assert.Equal(t, Secret{Name: "FOO", Value: "foobar"}, *secrets[1])

	err = os.RemoveAll(filepath.Join("testdata", envName))
	require.NoError(t, err)
	err = os.Setenv("OPENPAAS_PASSPHRASE", "")
	require.NoError(t, err)
}

func Test_Refresh_Add_One_Key_Remove_Another(t *testing.T) {
	err := os.Setenv("OPENPAAS_PASSPHRASE", "LongSecret")
	require.NoError(t, err)
	envName := RandStringBytes()
	err = os.MkdirAll(filepath.Join("testdata", envName, "secrets"), 0750)
	require.NoError(t, err)
	err = writePubKey("testdata", envName, johnDoeKey)
	require.NoError(t, err)
	err = writePubKey("testdata", envName, janeDoeKey)
	require.NoError(t, err)

	err = WriteSecret("testdata", envName, "FOO", "foobar")
	require.NoError(t, err)
	err = WriteSecret("testdata", envName, "BAR", "bazqux")
	require.NoError(t, err)

	secrets, err := getAllSecretsPrivate(johnArmored, "testdata", envName)
	require.NoError(t, err)
	assert.Len(t, secrets, 2)

	_, err = getAllSecretsPrivate(janeArmored, "testdata", envName)
	require.NoError(t, err)

	_, err = getAllSecretsPrivate(steveArmored, "testdata", envName)
	require.Error(t, err)

	err = writePubKey("testdata", envName, steveDoeKey)
	require.NoError(t, err)
	err = os.Remove(filepath.Join("testdata", envName, "pubkeys", "jane-doe-jane@doe.com.asc"))
	require.NoError(t, err)
	err = refreshPrivate(johnArmored, "testdata", envName)
	require.NoError(t, err)
	stevesSecrets, err := getAllSecretsPrivate(steveArmored, "testdata", envName)
	require.NoError(t, err)
	assert.Equal(t, stevesSecrets, secrets)

	_, err = getAllSecretsPrivate(janeArmored, "testdata", envName)
	require.Error(t, err)

	assert.Equal(t, Secret{Name: "BAR", Value: "bazqux"}, *secrets[0])

	assert.Equal(t, Secret{Name: "FOO", Value: "foobar"}, *secrets[1])

	err = os.RemoveAll(filepath.Join("testdata", envName))
	require.NoError(t, err)
	err = os.Setenv("OPENPAAS_PASSPHRASE", "")
	require.NoError(t, err)
}

func Test_Encrypt_No_Keys(t *testing.T) {
	err := os.Setenv("OPENPAAS_PASSPHRASE", "LongSecret")
	require.NoError(t, err)
	envName := RandStringBytes()
	err = os.MkdirAll(filepath.Join("testdata", envName, "secrets"), 0750)
	require.NoError(t, err)

	err = WriteSecret("testdata", envName, "FOO", "foobar")
	require.Error(t, err)

	err = os.RemoveAll(filepath.Join("testdata", envName))
	require.NoError(t, err)
}

func Test_Init_Env(t *testing.T) {
	envName := RandStringBytes()

	testFilesForEnv(t, envName, func() error {
		return initEnv("testdata", envName, johnDoeKey)
	})

}

func Test_Init_Secrets_With_No_Key_File(t *testing.T) {
	envName := RandStringBytes()

	err := os.RemoveAll(filepath.Join("testdata", ".openpaas"))
	require.NoError(t, err)
	init, err := initSecretsPrivate("testdata", "testdata", envName, func() (*keySettings, error) {
		return &keySettings{
			"john doe", "john@doe.com", "pass", "pass",
		}, nil
	})
	assert.True(t, init)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join("testdata", ".openpaas", "private-key.asc"))
	require.NoError(t, err)
	init, err = initSecretsPrivate("testdata", "testdata", envName, func() (*keySettings, error) {
		return nil, fmt.Errorf("should never be called")
	})

	assert.False(t, init)
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join("testdata", ".openpaas", "private-key.asc"))
	require.NoError(t, err)

	err = os.RemoveAll(filepath.Join("testdata", ".openpaas"))
	require.NoError(t, err)
}

func Test_Init_Secrets_With_Key_Set_In_Env(t *testing.T) {
	envName := RandStringBytes()

	err := os.Setenv("OPENPAAS_PRIVATE_KEY", johnArmored)
	require.NoError(t, err)

	testFilesForEnv(t, envName, func() error {
		_, e := InitSecrets("testdata", "testdata", envName)
		return e
	})
	err = os.Setenv("OPENPAAS_PRIVATE_KEY", "")
	require.NoError(t, err)
}

func testFilesForEnv(t *testing.T, env string, fn func() error) {
	err := os.RemoveAll(filepath.Join("testdata", env))
	require.NoError(t, err)
	err = fn()
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join("testdata", env, "pubkeys", "john-doe-john@doe.com.asc"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join("testdata", env, "secrets"))
	require.NoError(t, err)
	err = os.RemoveAll(filepath.Join("testdata", env))
	require.NoError(t, err)
}

func Test_RandString(t *testing.T) {
	pastStrings := []string{}

	for i := 1; i < 10; i++ {
		str := RandStringBytes()
		str = strings.ReplaceAll(str, " ", "")
		assert.GreaterOrEqual(t, len(str), 10)
		fmt.Println(str)
		assert.LessOrEqual(t, len(str), 30)

		for _, past := range pastStrings {
			assert.NotEqual(t, past, str)
		}
		pastStrings = append(pastStrings, str)
	}

}

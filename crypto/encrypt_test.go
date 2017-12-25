package crypto

import (
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		givenSecret, givenPlaintext string
	}{
		{
			givenSecret:    uuid.New().String(),
			givenPlaintext: "Lorem ipsum dolor sit amet, consectetur adipiscing elit",
		},
		{
			givenSecret:    uuid.New().String(),
			givenPlaintext: "Donec malesuada sapien nec turpis pellentesque lobortis. Suspendisse faucibus auctor eros eget auctor",
		},
		{
			givenSecret:    uuid.New().String(),
			givenPlaintext: "Ok",
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ciphertext, noncetext, err := EncryptWithSecretENV(tc.givenSecret, tc.givenPlaintext)
			assert.Nil(t, err)

			// repeats generate different ciphertext and noncetext
			ciphertext2, noncetext2, err := EncryptWithSecretENV(tc.givenSecret, tc.givenPlaintext)
			assert.Nil(t, err)
			assert.NotEqual(t, ciphertext, ciphertext2)
			assert.NotEqual(t, noncetext, noncetext2)

			// we can decrypt
			plaintext, err := DecryptWithSecretEnv(tc.givenSecret, ciphertext, noncetext)
			assert.Nil(t, err)
			assert.Equal(t, tc.givenPlaintext, plaintext)

			plaintext2, err := DecryptWithSecretEnv(tc.givenSecret, ciphertext2, noncetext2)
			assert.Nil(t, err)
			assert.Equal(t, tc.givenPlaintext, plaintext2)
		})
	}
}

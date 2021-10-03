package letsencrypt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"log"
	"os"

	"github.com/go-acme/lego/certcrypto"
	"github.com/go-acme/lego/certificate"
	"github.com/go-acme/lego/challenge/tlsalpn01"
	"github.com/go-acme/lego/lego"
	"github.com/go-acme/lego/registration"
)

var (
	ErrCertFilesAlreadyExist = errors.New("cert files already exist on disk")
)

type Certs struct {
	Domains        []string
	CertFilePath   string
	PrivateKeyPath string
	CADirectoryURL string
}

type User struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *User) GetEmail() string {
	return u.Email
}
func (u User) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *User) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

func (c Certs) FilesExist() (bool, error) {
	_, err := os.Stat(c.CertFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}

		return true, err
	}

	_, err = os.Stat(c.PrivateKeyPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return true, err
	}

	return true, nil

}

func (c Certs) RequestCerts(userEmail string) error {
	var (
		exist bool
		err   error
	)
	if exist, err = c.FilesExist(); err != nil {
		return err
	}
	if exist {
		return ErrCertFilesAlreadyExist
	}

	// Create a user. New accounts need an email and private key to start.
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	usr := &User{
		Email: userEmail,
		key:   privateKey,
	}

	config := lego.NewConfig(usr)

	config.CADirURL = c.CADirectoryURL
	config.Certificate.KeyType = certcrypto.RSA2048

	// A client facilitates communication with the CA server.
	client, err := lego.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	// Start listeners for ACME challenge
	// err = client.Challenge.SetHTTP01Provider(http01.NewProviderServer("", "8080"))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	cpSrv := tlsalpn01.NewProviderServer("", "8443")
	err = client.Challenge.SetTLSALPN01Provider(cpSrv)
	if err != nil {
		log.Fatal(err)
	}
	defer cpSrv.CleanUp("", "", "")

	// New users will need to register
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
	}
	usr.Registration = reg

	request := certificate.ObtainRequest{
		Domains: c.Domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
	}

	crtF, err := os.OpenFile(c.CertFilePath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer crtF.Close()
	_, err = crtF.Write(certificates.Certificate)
	if err != nil {
		log.Fatalln(err.Error())
	}

	pkF, err := os.OpenFile(c.PrivateKeyPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer pkF.Close()

	_, err = pkF.Write(certificates.PrivateKey)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return nil
}

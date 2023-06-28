package main

import (
	"crypto/tls"
	"crypto/x509"
	stderrors "errors"
	"flag"
	"github.com/guoyk93/rg"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssl "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205"
	"log"
	"os"
	"path/filepath"
)

const (
	DirData = "/data"
)

func newQcloudSSLClient(secretID, secretKey string) (*ssl.Client, error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "ssl.tencentcloudapi.com"
	return ssl.NewClient(common.NewCredential(secretID, secretKey), "", cpf)
}

func main() {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}()

	var (
		optDomain string
	)
	flag.StringVar(&optDomain, "domain", "", "domain name")
	flag.Parse()

	if optDomain == "" {
		err = stderrors.New("missing argument -domain")
		return
	}

	var (
		secretID  = os.Getenv("QCLOUD_SECRET_ID")
		secretKey = os.Getenv("QCLOUD_SECRET_KEY")
	)

	if secretID == "" {
		err = stderrors.New("missing environment variable QCLOUD_SECRET_ID")
		return
	}

	if secretKey == "" {
		err = stderrors.New("missing environment variable QCLOUD_SECRET_KEY")
		return
	}

	bufCrt := rg.Must(os.ReadFile(filepath.Join(DirData, optDomain, "fullchain.cer")))
	bufKey := rg.Must(os.ReadFile(filepath.Join(DirData, optDomain, optDomain+".key")))

	certT := rg.Must(tls.X509KeyPair(bufCrt, bufKey))
	certX := rg.Must(x509.ParseCertificate(certT.Certificate[0]))

	alias := certX.NotBefore.Format("20060102") + "-" + optDomain

	client := rg.Must(newQcloudSSLClient(secretID, secretKey))

	req := ssl.NewUploadCertificateRequest()
	req.Alias = common.StringPtr(alias)
	req.CertificatePublicKey = common.StringPtr(string(bufCrt))
	req.CertificatePrivateKey = common.StringPtr(string(bufKey))
	req.Repeatable = common.BoolPtr(false)

	if _, err = client.UploadCertificate(req); err != nil {
		if qErr, ok := err.(*errors.TencentCloudSDKError); ok {
			if qErr.Code == "FailedOperation.CertificateExists" {
				err = nil
			}
		}
	}
}

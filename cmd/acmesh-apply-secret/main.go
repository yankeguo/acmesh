package main

import (
	"bytes"
	"context"
	stderrors "errors"
	"flag"
	"github.com/guoyk93/rg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	LabelManagedBy = "app.kubernetes.io/managed-by"
	DirData        = "/data"
	NamespaceAll   = "_all"
)

func main() {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("exited with error:", err.Error())
	}()

	var (
		optDomain    string
		optNamespace string
		optName      string
	)

	flag.StringVar(&optDomain, "domain", "", "domain name of secret to apply")
	flag.StringVar(&optNamespace, "namespace", "", "namespace to apply secret, use '_all' for all namespaces")
	flag.StringVar(&optName, "name", "", "name of secret")
	flag.Parse()

	if optDomain == "" {
		err = stderrors.New("missing argument -domain")
		return
	}

	if optNamespace == "" {
		err = stderrors.New("missing argument -namespace")
		return
	}

	if optName == "" {
		err = stderrors.New("missing argument -secret")
		return
	}

	log.Println("loading certificate")

	bufCrt := rg.Must(os.ReadFile(filepath.Join(DirData, optDomain, "fullchain.cer")))
	bufKey := rg.Must(os.ReadFile(filepath.Join(DirData, optDomain, optDomain+".key")))

	log.Println("certificate loaded for:", optDomain)

	ctx := context.Background()

	client := rg.Must(kubernetes.NewForConfig(rg.Must(clientcmd.DefaultClientConfig.ClientConfig())))

	var namespaces []string

	if optNamespace == NamespaceAll {
		for _, item := range rg.Must(client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})).Items {
			namespaces = append(namespaces, item.Name)
		}
	} else {
		namespaces = []string{optNamespace}
	}

	log.Println("namespaces to updated:", strings.Join(namespaces, ", "))

	for _, namespace := range namespaces {
		log.Println("applying:", namespace)

		var current *corev1.Secret

		if current, err = client.CoreV1().Secrets(namespace).Get(ctx, optName, metav1.GetOptions{}); err != nil {
			if errors.IsNotFound(err) {
				err = nil
			} else {
				return
			}
		}

		if current != nil {
			if bytes.Equal(current.Data[corev1.TLSCertKey], bufCrt) &&
				bytes.Equal(current.Data[corev1.TLSPrivateKeyKey], bufKey) {
				log.Println("  already up to date")
				continue
			}

			if current.Labels == nil {
				current.Labels = map[string]string{}
			}
			current.Labels[LabelManagedBy] = "acmesh-apply-secret"
			current.Data[corev1.TLSCertKey] = bufCrt
			current.Data[corev1.TLSPrivateKeyKey] = bufKey

			rg.Must(client.CoreV1().Secrets(namespace).Update(ctx, current, metav1.UpdateOptions{}))

			log.Println("  updated")
		} else {
			rg.Must(client.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      optName,
					Namespace: namespace,
				},
				Type: corev1.SecretTypeTLS,
				Data: map[string][]byte{
					corev1.TLSCertKey:       bufCrt,
					corev1.TLSPrivateKeyKey: bufKey,
				},
			}, metav1.CreateOptions{}))

			log.Println("  created")
		}
	}

	log.Println("all done")
}

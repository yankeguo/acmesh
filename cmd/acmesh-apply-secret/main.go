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
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	LabelManagedByKey   = "app.kubernetes.io/managed-by"
	LabelManagedByValue = "acmesh-apply-secret"
	DirData             = "/data"
	NamespaceAll        = "_all"
)

func main() {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}()
	defer rg.Guard(&err)

	var (
		optKubeconfig string
		optDomain     string
		optNamespace  string
		optName       string
	)

	flag.StringVar(&optKubeconfig, "kubeconfig", os.Getenv("KUBECONFIG"), "kubeconfig file")
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

	log.Println("loading certificate:", optDomain)

	bufCrt := rg.Must(os.ReadFile(filepath.Join(DirData, optDomain, "fullchain.cer")))
	bufKey := rg.Must(os.ReadFile(filepath.Join(DirData, optDomain, optDomain+".key")))

	ctx := context.Background()

	if optKubeconfig == "" {
		optKubeconfig = filepath.Join(homedir.HomeDir(), ".kube", "config")

		if _, err = os.Stat(optKubeconfig); err != nil {
			if os.IsNotExist(err) {
				err = nil
				optKubeconfig = ""
			} else {
				return
			}
		}
	}

	var config *rest.Config

	if optKubeconfig == "" {
		config = rg.Must(rest.InClusterConfig())
	} else {
		config = rg.Must(clientcmd.BuildConfigFromFlags("", optKubeconfig))
	}

	client := rg.Must(kubernetes.NewForConfig(config))

	var namespaces []string

	if optNamespace == NamespaceAll {
		for _, item := range rg.Must(client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})).Items {
			namespaces = append(namespaces, item.Name)
		}
	} else {
		namespaces = strings.Split(optNamespace, ",")
		for i := range namespaces {
			namespaces[i] = strings.TrimSpace(namespaces[i])
		}
	}

	log.Println("namespaces:", strings.Join(namespaces, ", "))

	for _, namespace := range namespaces {
		log.Println("working:", namespace)

		var current *corev1.Secret

		if current, err = client.CoreV1().Secrets(namespace).Get(ctx, optName, metav1.GetOptions{}); err != nil {
			if errors.IsNotFound(err) {
				err = nil

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
			} else {
				return
			}
		} else {
			if current.Labels == nil {
				current.Labels = map[string]string{}
			}
			if current.Data == nil {
				current.Data = map[string][]byte{}
			}

			if bytes.Equal(current.Data[corev1.TLSCertKey], bufCrt) &&
				bytes.Equal(current.Data[corev1.TLSPrivateKeyKey], bufKey) &&
				current.Labels[LabelManagedByKey] == LabelManagedByValue {
				log.Println("  already up to date")
				continue
			}

			current.Data[corev1.TLSCertKey] = bufCrt
			current.Data[corev1.TLSPrivateKeyKey] = bufKey
			current.Labels[LabelManagedByKey] = LabelManagedByValue

			rg.Must(client.CoreV1().Secrets(namespace).Update(ctx, current, metav1.UpdateOptions{}))

			log.Println("  updated")
		}
	}

	log.Println("all done")
}

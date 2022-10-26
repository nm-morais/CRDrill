package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/gertd/go-pluralize"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var (
	NotFoundErr = errors.New("the server could not find the requested resource")
	logger      = log.New()
	pluralizeCl = pluralize.NewClient()
)

func getCRD(clientset *kubernetes.Clientset, crdPath string, crd *CrossplaneCRD) error {
	// logger.Info("Querying:", crdPath)
	d, err := clientset.RESTClient().Get().AbsPath(crdPath).DoRaw(context.TODO())
	if err != nil {
		return err

	}
	if err := json.Unmarshal(d, &crd); err != nil {
		logger.Panic(err)
	}
	return nil
}

func findNonReadySubResources(clientSet *kubernetes.Clientset, crd CrossplaneCRD) {
	for _, ref := range crd.Spec.ResourceRefs {
		crdKindLower := strings.ToLower(ref.Kind)
		crdKindPLural := pluralizeCl.Plural(crdKindLower)
		crdPath := fmt.Sprintf("/apis/%s/%s/%s", ref.ApiVersion, crdKindPLural, ref.Name)
		childCRD := CrossplaneCRD{}

		if ref.Name == "" {
			log.Warnf("Subresource of %s, name: %s, kind: %s has not been created", ref.ApiVersion, ref.Kind, ref.Name)
			continue
		}

		if err := getCRD(clientSet, crdPath, &childCRD); err != nil {
			log.Errorf("Got error: (%s) on CRD of type: %s/%s, name: %s", err, ref.ApiVersion, crdKindPLural, ref.Name)
			continue
		}

		ready, err := childCRD.IsReady()

		if err != nil {
			if err == NoStatusConditionErr && childCRD.Kind == "ProviderConfig" {
				// providers never have a "ready" status
				continue
			}
			log.Warnf("%v has error: %s", childCRD, err)
			continue
		}

		if err == nil && !ready { // has "Ready" condition and is not ready
			log.Warnf("%v is not ready", childCRD)
			findNonReadySubResources(clientSet, childCRD) // recursive step
		}

		if ok, err := childCRD.HasReconcileError(); ok {
			log.Errorf("%v has reconcileError: %s", childCRD, err)
			findNonReadySubResources(clientSet, childCRD) // recursive step
		}

	}
}

func CRDrill(clientSet *kubernetes.Clientset, kubeConfig string, crdType string, crdName string) {
	crdPath := fmt.Sprintf("/apis/websummit.com/v1beta1/%s/%s", crdType, crdName)
	logger.Info()
	logger.Infof("Using kubeconfig: %s", kubeConfig)
	logger.Infof("Looking for CRD of type: %s, on path: %s", crdType, crdPath)
	logger.Info()

	crd := CrossplaneCRD{}
	if err := getCRD(clientSet, crdPath, &crd); err != nil {
		log.Panicf("Got error: %s. Obtaining CRD on path: %s", err, crdPath)
	}
	log.Info("Inspecting CRD: ", crd)
	ready, _ := crd.IsReady()
	if !ready { // not ready
		findNonReadySubResources(clientSet, crd) // recursive step
	} else {
		log.Infof("%v is ready", crd)
	}
	logger.Info()
	logger.Info("All done.")
}

// returns: (kubeconfig|crdType|crdName)
func parseArgs() (*string, *string, *string) {

	var kubeconfig *string

	if os.Getenv("KUBECONFIG") != "" {
		aux := os.Getenv("KUBECONFIG")
		kubeconfig = &aux
	} else {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		} else {
			kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
		}
	}

	crdType := flag.String("type", "", "type of resource (plural), examples: platforms|tenants|eksclusters|....")
	crdName := flag.String("name", "", "name of resource")
	flag.Parse()

	if crdType == nil {
		fmt.Println("argumment 'type' is null")
		os.Exit(1)
	}
	if crdName == nil {
		fmt.Println("argumment 'name' is null")
		os.Exit(1)
	}

	if *crdType == "" {
		fmt.Println("argumment 'type' is empty")
		os.Exit(1)
	}
	if *crdName == "" {
		fmt.Println("argumment 'name' is empty")
		os.Exit(1)
	}

	return kubeconfig, crdType, crdName

}

func main() {

	kubeConfig, crdType, crdName := parseArgs()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientSet
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	CRDrill(clientSet, *kubeConfig, *crdType, *crdName)
}

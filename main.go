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
	NoStatusConditionErr = errors.New("CRD does not have a status condition")
	NotFoundErr          = errors.New("the server could not find the requested resource")
	logger               = log.New()
	pluralizeCl          = pluralize.NewClient()
)

func getCRD(clientset *kubernetes.Clientset, crdPath string, crd *CrossplaneCRD) {
	// logger.Info("Querying:", crdPath)
	d, err := clientset.RESTClient().Get().AbsPath(crdPath).DoRaw(context.TODO())
	if err != nil {
		if err.Error() == NotFoundErr.Error() {
			logger.Panicf("Got err: %s, verify if your specified type is correct, and if you are specifying the correct context", err)
		}
		logger.Panicf("Got err: %s", err)
	}
	if err := json.Unmarshal(d, &crd); err != nil {
		logger.Panic(err)
	}
}

func findNonReadySubResources(clientSet *kubernetes.Clientset, crd CrossplaneCRD) {
	for _, ref := range crd.Spec.ResourceRefs {
		crdKindLower := strings.ToLower(ref.Kind)
		crdKindPLural := pluralizeCl.Plural(crdKindLower)
		crdPath := fmt.Sprintf("/apis/%s/%s/%s", ref.ApiVersion, crdKindPLural, ref.Name)
		childCRD := CrossplaneCRD{}
		getCRD(clientSet, crdPath, &childCRD)
		ready, err := childCRD.IsReady()
		if err == nil && !ready { // has "Ready" condition and is not ready
			log.Warnf("%v is not ready", childCRD)
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
	getCRD(clientSet, crdPath, &crd)
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

//  returns: (kubeconfig|crdType|crdName)
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

	if *crdType == "" {
		logger.Panic("argumment 'type' is null")
	}
	if *crdName == "" {
		logger.Panic("argumment 'name' is null")
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

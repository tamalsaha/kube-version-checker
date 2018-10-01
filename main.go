package main

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go-version"
	"github.com/appscode/go/log"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	masterURL := ""
	kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")

	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		log.Fatalf("Could not get Kubernetes config: %s", err)
	}

	kc := kubernetes.NewForConfigOrDie(config)

	info, err := kc.Discovery().ServerVersion()
	if err != nil {
		log.Fatalln(err)
	}
	gv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		log.Fatalln(err)
	}
	v2 := gv.ToMutator().ResetMetadata().ResetPrerelease().Done()
	fmt.Println(v2)

	c, err := version.NewConstraint(">= 1.11.0")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(c.String())

	fmt.Println("MATCHES = ", c.Check(v2))
}

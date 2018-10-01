package main

import (
	"fmt"
	"path/filepath"

	"github.com/appscode/go-version"
	"github.com/appscode/go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
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

	utilruntime.Must(IsDefaultSupportedVersion(kc))
}

var (
	defaultConstraint = ">= 1.9.0"
	defaultBlackListedVersions = []string{"1.11.0", "1.11.1", "1.11.2"}
	defaultBlackListedMultiMasterVersions = []string{
		"1.9.0", "1.9.1", "1.9.2", "1.9.3", "1.9.4", "1.9.5", "1.9.6", "1.9.7",
		"1.10.0", "1.10.1",
	}
)

func IsDefaultSupportedVersion(kc kubernetes.Interface) error {
	return IsSupportedVersion(
		kc,
		defaultConstraint,
		defaultBlackListedVersions,
		defaultBlackListedMultiMasterVersions)
}

func IsSupportedVersion(kc kubernetes.Interface, constraint string, blackListedVersions []string, blackListedMultiMasterVersions []string) error {
	info, err := kc.Discovery().ServerVersion()
	if err != nil {
		return err
	}
	gv, err := version.NewVersion(info.GitVersion)
	if err != nil {
		return err
	}
	v := gv.ToMutator().ResetMetadata().ResetPrerelease().Done()

	nodes, err := kc.CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/master",
	})
	if err != nil {
		return err
	}
	multiMaster := len(nodes.Items) > 1

	return checkVersion(v, multiMaster, constraint, blackListedVersions, blackListedMultiMasterVersions)
}

func checkVersion(v *version.Version, multiMaster bool, constraint string, blackListedVersions []string, blackListedMultiMasterVersions []string) error {
	vs := v.String()

	if constraint != "" {
		c, err := version.NewConstraint(constraint)
		if err != nil {
			return err
		}
		if !c.Check(v) {
			return fmt.Errorf("kubernetes version %s fails constraint %s", vs, constraint)
		}
	}

	if len(blackListedVersions) > 0 {
		list := sets.NewString(blackListedVersions...)
		if list.Has(v.Original()) {
			return fmt.Errorf("kubernetes version %s is blacklisted", v.Original())
		}
		if list.Has(vs) {
			return fmt.Errorf("kubernetes version %s is blacklisted", vs)
		}
	}

	if len(blackListedMultiMasterVersions) > 0 && multiMaster {
		list := sets.NewString(blackListedMultiMasterVersions...)
		if list.Has(v.Original()) {
			return fmt.Errorf("kubernetes version %s is blacklisted for multi-master cluster", v.Original())
		}
		if list.Has(vs) {
			return fmt.Errorf("kubernetes version %s is blacklisted for multi-master cluster", vs)
		}
	}
	return nil
}

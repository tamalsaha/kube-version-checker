package main

import (
	"fmt"
	"github.com/pkg/errors"
	"path/filepath"

	"github.com/appscode/go-version"
	"github.com/appscode/go/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
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

type KnownBug struct {
	BugURL string
	Fix    string
}

func (e *KnownBug) Error() string {
	return "Bug: " + e.BugURL + ". To fix, " + e.Fix
}

var err62649_K1_9 = &KnownBug{BugURL: "https://github.com/kubernetes/kubernetes/pull/62649", Fix: "upgrade to Kubernetes 1.9.8 or later"}
var err62649_K1_10 = &KnownBug{BugURL: "https://github.com/kubernetes/kubernetes/pull/62649", Fix: "upgrade to Kubernetes 1.10.2 or later"}

var (
	defaultConstraint                     = ">= 1.9.0"
	defaultBlackListedVersions            map[string]error
	defaultBlackListedMultiMasterVersions = map[string]error{
		"1.9.0":  err62649_K1_9,
		"1.9.1":  err62649_K1_9,
		"1.9.2":  err62649_K1_9,
		"1.9.3":  err62649_K1_9,
		"1.9.4":  err62649_K1_9,
		"1.9.5":  err62649_K1_9,
		"1.9.6":  err62649_K1_9,
		"1.9.7":  err62649_K1_9,
		"1.10.0": err62649_K1_10,
		"1.10.1": err62649_K1_10,
	}
)

func IsDefaultSupportedVersion(kc kubernetes.Interface) error {
	return IsSupportedVersion(
		kc,
		defaultConstraint,
		defaultBlackListedVersions,
		defaultBlackListedMultiMasterVersions)
}

func IsSupportedVersion(kc kubernetes.Interface, constraint string, blackListedVersions map[string]error, blackListedMultiMasterVersions map[string]error) error {
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

func checkVersion(v *version.Version, multiMaster bool, constraint string, blackListedVersions map[string]error, blackListedMultiMasterVersions map[string]error) error {
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

	if e, ok := blackListedVersions[v.Original()]; ok {
		return errors.Wrapf(e, "kubernetes version %s is blacklisted", v.Original())
	}
	if e, ok := blackListedVersions[vs]; ok {
		return errors.Wrapf(e, "kubernetes version %s is blacklisted", vs)
	}

	if multiMaster {
		if e, ok := blackListedMultiMasterVersions[v.Original()]; ok {
			return errors.Wrapf(e, "kubernetes version %s is blacklisted for multi-master cluster", v.Original())
		}
		if e, ok := blackListedMultiMasterVersions[vs]; ok {
			return errors.Wrapf(e, "kubernetes version %s is blacklisted for multi-master cluster", vs)
		}
	}
	return nil
}

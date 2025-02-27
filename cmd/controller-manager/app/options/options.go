/*
Copyright 2020 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package options

import (
	"flag"
	"strings"
	"time"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/spf13/pflag"
	"k8s.io/client-go/tools/leaderelection"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/simple/client/devops/jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
	ldapclient "kubesphere.io/kubesphere/pkg/simple/client/ldap"
	"kubesphere.io/kubesphere/pkg/simple/client/multicluster"
	"kubesphere.io/kubesphere/pkg/simple/client/network"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/simple/client/s3"
	"kubesphere.io/kubesphere/pkg/simple/client/servicemesh"
)

type KubeSphereControllerManagerOptions struct {
	KubernetesOptions     *k8s.KubernetesOptions
	DevopsOptions         *jenkins.Options
	S3Options             *s3.Options
	AuthenticationOptions *authentication.Options
	LdapOptions           *ldapclient.Options
	OpenPitrixOptions     *openpitrix.Options
	NetworkOptions        *network.Options
	MultiClusterOptions   *multicluster.Options
	ServiceMeshOptions    *servicemesh.Options
	GatewayOptions        *gateway.Options
	LeaderElect           bool
	LeaderElection        *leaderelection.LeaderElectionConfig
	WebhookCertDir        string

	// KubeSphere is using sigs.k8s.io/application as fundamental object to implement Application Management.
	// There are other projects also built on sigs.k8s.io/application, when KubeSphere installed along side
	// them, conflicts happen. So we leave an option to only reconcile applications  matched with the given
	// selector. Default will reconcile all applications.
	//    For example
	//      "kubesphere.io/creator=" means reconcile applications with this label key
	//      "!kubesphere.io/creator" means exclude applications with this key
	ApplicationSelector string
}

func NewKubeSphereControllerManagerOptions() *KubeSphereControllerManagerOptions {
	s := &KubeSphereControllerManagerOptions{
		KubernetesOptions:     k8s.NewKubernetesOptions(),
		DevopsOptions:         jenkins.NewDevopsOptions(),
		S3Options:             s3.NewS3Options(),
		LdapOptions:           ldapclient.NewOptions(),
		OpenPitrixOptions:     openpitrix.NewOptions(),
		NetworkOptions:        network.NewNetworkOptions(),
		MultiClusterOptions:   multicluster.NewOptions(),
		ServiceMeshOptions:    servicemesh.NewServiceMeshOptions(),
		AuthenticationOptions: authentication.NewOptions(),
		GatewayOptions:        gateway.NewGatewayOptions(),
		LeaderElection: &leaderelection.LeaderElectionConfig{
			LeaseDuration: 30 * time.Second,
			RenewDeadline: 15 * time.Second,
			RetryPeriod:   5 * time.Second,
		},
		LeaderElect:         false,
		WebhookCertDir:      "",
		ApplicationSelector: "",
	}

	return s
}

func (s *KubeSphereControllerManagerOptions) Flags() cliflag.NamedFlagSets {
	fss := cliflag.NamedFlagSets{}

	s.KubernetesOptions.AddFlags(fss.FlagSet("kubernetes"), s.KubernetesOptions)
	s.DevopsOptions.AddFlags(fss.FlagSet("devops"), s.DevopsOptions)
	s.S3Options.AddFlags(fss.FlagSet("s3"), s.S3Options)
	s.AuthenticationOptions.AddFlags(fss.FlagSet("authentication"), s.AuthenticationOptions)
	s.LdapOptions.AddFlags(fss.FlagSet("ldap"), s.LdapOptions)
	s.OpenPitrixOptions.AddFlags(fss.FlagSet("openpitrix"), s.OpenPitrixOptions)
	s.NetworkOptions.AddFlags(fss.FlagSet("network"), s.NetworkOptions)
	s.MultiClusterOptions.AddFlags(fss.FlagSet("multicluster"), s.MultiClusterOptions)
	s.ServiceMeshOptions.AddFlags(fss.FlagSet("servicemesh"), s.ServiceMeshOptions)
	s.GatewayOptions.AddFlags(fss.FlagSet("gateway"), s.GatewayOptions)
	fs := fss.FlagSet("leaderelection")
	s.bindLeaderElectionFlags(s.LeaderElection, fs)

	fs.BoolVar(&s.LeaderElect, "leader-elect", s.LeaderElect, ""+
		"Whether to enable leader election. This field should be enabled when controller manager"+
		"deployed with multiple replicas.")

	fs.StringVar(&s.WebhookCertDir, "webhook-cert-dir", s.WebhookCertDir, ""+
		"Certificate directory used to setup webhooks, need tls.crt and tls.key placed inside."+
		"if not set, webhook server would look up the server key and certificate in"+
		"{TempDir}/k8s-webhook-server/serving-certs")

	gfs := fss.FlagSet("generic")
	gfs.StringVar(&s.ApplicationSelector, "application-selector", s.ApplicationSelector, ""+
		"Only reconcile application(sigs.k8s.io/application) objects match given selector, this could avoid conflicts with "+
		"other projects built on top of sig-application. Default behavior is to reconcile all of application objects.")

	kfs := fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		kfs.AddGoFlag(fl)
	})

	return fss
}

func (s *KubeSphereControllerManagerOptions) Validate() []error {
	var errs []error
	errs = append(errs, s.DevopsOptions.Validate()...)
	errs = append(errs, s.KubernetesOptions.Validate()...)
	errs = append(errs, s.S3Options.Validate()...)
	errs = append(errs, s.OpenPitrixOptions.Validate()...)
	errs = append(errs, s.NetworkOptions.Validate()...)
	errs = append(errs, s.LdapOptions.Validate()...)
	errs = append(errs, s.MultiClusterOptions.Validate()...)

	if len(s.ApplicationSelector) != 0 {
		_, err := labels.Parse(s.ApplicationSelector)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (s *KubeSphereControllerManagerOptions) bindLeaderElectionFlags(l *leaderelection.LeaderElectionConfig, fs *pflag.FlagSet) {
	fs.DurationVar(&l.LeaseDuration, "leader-elect-lease-duration", l.LeaseDuration, ""+
		"The duration that non-leader candidates will wait after observing a leadership "+
		"renewal until attempting to acquire leadership of a led but unrenewed leader "+
		"slot. This is effectively the maximum duration that a leader can be stopped "+
		"before it is replaced by another candidate. This is only applicable if leader "+
		"election is enabled.")
	fs.DurationVar(&l.RenewDeadline, "leader-elect-renew-deadline", l.RenewDeadline, ""+
		"The interval between attempts by the acting master to renew a leadership slot "+
		"before it stops leading. This must be less than or equal to the lease duration. "+
		"This is only applicable if leader election is enabled.")
	fs.DurationVar(&l.RetryPeriod, "leader-elect-retry-period", l.RetryPeriod, ""+
		"The duration the clients should wait between attempting acquisition and renewal "+
		"of a leadership. This is only applicable if leader election is enabled.")
}

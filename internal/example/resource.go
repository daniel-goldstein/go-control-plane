// Copyright 2020 Envoyproxy Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package example

import (
	"fmt"
	// "strings"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

const (
	ClusterName  = "example_proxy_cluster"
	RouteName    = "local_route"
	ListenerName = "listener_0"
	ListenerPort = 10000
	UpstreamHost = "www.envoyproxy.io"
	UpstreamPort = 80
	SecretName   = "server_cert"

	serverCert = `-----BEGIN CERTIFICATE-----
MIIEcjCCAlqgAwIBAgIUQfxVKIOdD4v3BO8MVpIFCWZnewswDQYJKoZIhvcNAQEL
BQAwEDEOMAwGA1UEAwwFTXkgQ0EwHhcNMjYwMzE4MTQ1OTQzWhcNMjcwMzE4MTQ1
OTQzWjAVMRMwEQYDVQQDDApteS1zZXJ2aWNlMIIBIjANBgkqhkiG9w0BAQEFAAOC
AQ8AMIIBCgKCAQEAwFV+0HNkOD2v9C8J4EN1lXvCrEuc3QtScYtEsNX8jozK6Yac
M4gu8vs6zuogiirKD/0ZrOfLhIKdyiMwzwKFXTjDTOWjMeix8cKohNp+1tKjd+Zt
kI22uqz2M26oLEzUojZxuntJ2aHB9vfpXg9+626/LiLnlUYvhwVlVIu0ZUI+Nph+
/x1+JdfFU3g5OSjp3nSPtGo+XTHRvA6yd6OG/607n2f2/w+WLvwxFn6mih0dBguJ
zC41L/tSzZ9NzzbLkiX3f/NOWWmCEY4Z4jz4E+LBmJGz8mmCjQBw+MDd5wlYn6G1
wMtfPZMKzTXZDgqK8fM5EewK4B/xMhHncgb63QIDAQABo4G+MIG7MB8GA1UdIwQY
MBaAFOmWs3qzS511RSDG6zpgsmGUGF9EMAkGA1UdEwQCMAAwCwYDVR0PBAQDAgWg
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMEwGA1UdEQRFMEOCCm15LXNlcnZpY2WCJG15
LXNlcnZpY2UuZGVmYXVsdC5zdmMuY2x1c3Rlci5sb2NhbIIJbG9jYWxob3N0hwR/
AAABMB0GA1UdDgQWBBT21Cpxr7Kzq0Oqd7A0pq7HPIYKkjANBgkqhkiG9w0BAQsF
AAOCAgEAhMqMDHBIGSwKz8jITXokU7LRhnIt/boL8b0s1v9QSluvb02L3FK1LOiB
LmiWSvanYn5OthKxyVSbl+Oa46jWBAiobJ5bMi1sO+AiP8rz8i5vVSY3czFDRvW5
5AfKfpCSS7lcqjt9CFsbj04iMmYhzOHYz8Gk/S3qBVvpvxFr/8L+5uKZMuT2XVNj
VnOJ7w7G41qBlX1pInuRVBTo5MnZPXSoyaHSbOeAOvrxbSoSQUp7IQVPKP67+Lvq
FDrr6cpzQtvCsTsuQzrnc/jH2rl2DgcT4iSeb4MKjAU/6g0OaXKpjlsd+mLIF66D
L7hLF96WCDirkzTHvMVMElu0PI/0Gi8tqU6IDX9tfbuDUTbFkxSAGr/WL4WjoWfu
1TAyyfmtxMVYseP4U8rJD0osfswN0SjKrQtqiLACe0FScFFBAK7vx6r7hHP6I45q
yOi84s8Tg6h6a2a/OQGaz7wRr9FyZlj8jKFHAjhD5EVnwc+eeZptPRjZmH1NrShV
9Ok/w/ioQD+hVcsAbzOrLgRF+MzkWCsYooWpH7pd3UfDsMC8R2rqU1vqp7M1d3ya
uPfJHiHxG9yKzt/b47HKrl7jJVS/GTdGG9oSPdYNDBUXyRz7+phupGFRfH46m06A
OQxOdETIRKtOPCyxLrpcWcJKsQjIGZnWKh9f9VfNJmcxg+fWa9M=
-----END CERTIFICATE-----
`

	serverKey = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDAVX7Qc2Q4Pa/0
LwngQ3WVe8KsS5zdC1Jxi0Sw1fyOjMrphpwziC7y+zrO6iCKKsoP/Rms58uEgp3K
IzDPAoVdOMNM5aMx6LHxwqiE2n7W0qN35m2Qjba6rPYzbqgsTNSiNnG6e0nZocH2
9+leD37rbr8uIueVRi+HBWVUi7RlQj42mH7/HX4l18VTeDk5KOnedI+0aj5dMdG8
DrJ3o4b/rTufZ/b/D5Yu/DEWfqaKHR0GC4nMLjUv+1LNn03PNsuSJfd/805ZaYIR
jhniPPgT4sGYkbPyaYKNAHD4wN3nCVifobXAy189kwrNNdkOCorx8zkR7ArgH/Ey
EedyBvrdAgMBAAECggEACj0S4uRdp6vK/9f4Md7Ndd9wec0NpOu7IA2+op5Fk04V
6DARBSJBA5DRrN2kTU6hUpAR65UsTZnJbg8JBGAZOuDwbpnD4f6F6H2JDId4HJX6
e0HCP+F3YoEeGxdPLwqJADifPcLd5863XV8NpoLzKfPLjBhyFQ13LrwgDIviVsiD
4yPC3uDi07yOaGjM8Tiry6jH7fzhotY28Vzqe/exMFVoyu0xml1TgpS/OWTT/gsw
Ak7dpipqgTyD8nAejn/rfv93Jexf5UzgUUDO8Nicjv50S8l8itdMl8qwOvPxmHL6
uFpu/70vqnHoqlfNh7OAQsm62jxnFuyJ5rq5tBVwOQKBgQD9gzUEZvwIMUWKFwNx
9dnB/e3eORkYpkYdXYxqd7RWiDaAnlxrSOQIuIXPxQ7iWQckngdvupJAKFnOzD/e
dTCk8w02Znpq27aD77oHhwz0XJ+BAMe+NdCR87+DM6Tgh/r504F/wHm2rDMCkjmJ
OTEDOq1+o1UMY1w4LrfVR1XoNQKBgQDCOJ13NWYu5b0qaqav0RryEPNS7o3EdBC4
0jzNhVedNkcnSVL6TrFnkcbyKEkfWxmsDls+APKMxAwx4TD8lB+Tr8ZbjTBUkrh4
GK77naWANqWv2qJwjHRKwY35jiLKxCXuSIvnsYiIzblAqJQ7wxb801Jy+2KF1Crj
PcCDOH6tCQKBgQDRnqJCB52ycHtdmXXhzzXFsF/1diUIOsSTF305s81MF8lpRIiK
tXTIuTr796c9BfxgDMN9YTn5DuRjmIPfP+t/GPH933Kt1QrvwVODUeomTEgfdTO0
Ve8mH/RlWlikyAuAc6EKr25026I6KAqnKsEaOHSo2AlE+wuP8SFUm22vWQKBgAj9
yfxs0nA1Xo6KJXFaQt8V/c3HEXUY0nVb9kildarnil+9O0QvRHNBAm7PgqMa+pNG
jt7N+Gyf3tioTjZDPTr/FjXC0Yv4xuV4bxFi+Ph4jy8W9hIzzmZvk30MIXw1nHPt
k9yEEYgTzhG6PDKQE45c0iJUlPkRG3Mttq3cfbDRAoGBAPDcOV/axxeHKabyj6h1
PKLuUuzjndSms+WuOJFETogo7DyxUHT3FSvD+KWWKJ9AQTwmX4xyoza4jU+KxAVC
thJCPALfzgV15j4uLln2x4kd32wGGoxoIaSC/QKxpdaBBut2sirAGfnUrHPP1NuC
BeKqMI+IY9HMmBOKciGuAbhP
-----END PRIVATE KEY-----
`
)

var (
	countOfWhitespace = 0
)

func makeCluster(clusterName string) *cluster.Cluster {
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       makeEndpoint(clusterName),
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
	}
}

func makeEndpoint(clusterName string) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.SocketAddress_TCP,
									Address:  UpstreamHost,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: UpstreamPort,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func makeRoute(routeName, clusterName string) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "local_service",
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
						HostRewriteSpecifier: &route.RouteAction_HostRewriteLiteral{
							HostRewriteLiteral: UpstreamHost,
						},
					},
				},
			}},
		}},
	}
}

func makeHTTPListener(listenerName, route, secretName string, initialFetchTimeout time.Duration) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    makeConfigSource(initialFetchTimeout),
				RouteConfigName: route,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       "http-router",
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}
	pbst, err := anypb.New(manager)
	if err != nil {
		panic(err)
	}

	downstreamTLSContext, err := anypb.New(&tlsv3.DownstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*tlsv3.SdsSecretConfig{{
				Name:      secretName,
				SdsConfig: makeConfigSource(initialFetchTimeout),
			}},
		},
	})
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "0.0.0.0",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: ListenerPort,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			TransportSocket: &core.TransportSocket{
				Name: "envoy.transport_sockets.tls",
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: downstreamTLSContext,
				},
			},
			Filters: []*listener.Filter{{
				Name: "http-connection-manager",
				ConfigType: &listener.Filter_TypedConfig{
					TypedConfig: pbst,
				},
			}},
		}},
	}
}

func makeSecret(secretName string) *tlsv3.Secret {
	// Mutating the secret also resolvs the failure because the
	// resource version no longer match what Envoy reports to have
	// key := serverKey + strings.Repeat(" ", countOfWhitespace)
	// countOfWhitespace += 1
	key := serverKey
	return &tlsv3.Secret{
		Name: secretName,
		Type: &tlsv3.Secret_TlsCertificate{
			TlsCertificate: &tlsv3.TlsCertificate{
				CertificateChain: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: serverCert},
				},
				PrivateKey: &core.DataSource{
					Specifier: &core.DataSource_InlineString{InlineString: key},
				},
			},
		},
	}
}

func makeConfigSource(initialFetchTimeout time.Duration) *core.ConfigSource {
	source := &core.ConfigSource{
		InitialFetchTimeout: durationpb.New(initialFetchTimeout),
	}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	// Using a dedicated ApiConfigSource for the secret works, presumably because
	// the new stream has different semantics than joining the existing ADS stream
	// which remembers the resource version that Envoy previously held
	// source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
	// 	ApiConfigSource: &core.ApiConfigSource{
	// 		TransportApiVersion:       resource.DefaultAPIVersion,
	// 		ApiType:                   core.ApiConfigSource_GRPC,
	// 		SetNodeOnFirstMessageOnly: true,
	// 		GrpcServices: []*core.GrpcService{{
	// 			TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
	// 				EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: "xds_cluster"},
	// 			},
	// 		}},
	// 	},
	// }
	source.ConfigSourceSpecifier = &core.ConfigSource_Ads{
		Ads: &core.AggregatedConfigSource{},
	}
	return source
}

func GenerateSnapshot(initialFetchTimeout time.Duration) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(fmt.Sprintf("%d", time.Now().UnixMilli()),
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {makeCluster(ClusterName)},
			resource.RouteType:    {makeRoute(RouteName, ClusterName)},
			resource.ListenerType: {makeHTTPListener(ListenerName, RouteName, SecretName, initialFetchTimeout)},
			resource.SecretType:   {makeSecret(SecretName)},
		},
	)
	return snap
}

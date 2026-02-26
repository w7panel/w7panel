#!/bin/sh

kubectl apply -f - <<EOF
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
    name: higress-gateway-global-route-config
    namespace: higress-system
spec:
    configPatches:
        -
            applyTo: NETWORK_FILTER
            match:
                context: GATEWAY
                listener:
                    filterChain:
                        filter:
                            name: envoy.filters.network.http_connection_manager
            patch:
                operation: MERGE
                value:
                    typed_config:
                        '@type': >-
                            type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                        skip_xff_append: true
                        xff_num_trusted_hops: 2
        -
            applyTo: ROUTE_CONFIGURATION
            match:
                context: GATEWAY
            patch:
                operation: MERGE
                value:
                    request_headers_to_add:
                        -
                            append: false
                            header:
                                key: x-real-ip
                                value: '%REQ(X-Forwarded-For)%'
                        -
                            append: false
                            header:
                                key: X-Forwarded-Proto
                                value: '%REQ(X-Forwarded-Proto)%'
EOF


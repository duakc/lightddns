// Package ddnsmetric is the bridge between the generic infra/metrics
// Registry and the lightddns metric naming convention.
//
// Every metric exposed by lightddns lives under the "ddns" namespace and a
// per-variant subsystem ("domain"/"provider"/"datasource"/"service"). A
// Factory pre-binds namespace + subsystem so callers only specify the leaf
// metric name. The final metric name is built with prometheus.BuildFQName.
//
// Variant impls (Provider/Datasource/Service) obtain their Factory from
// FromContext(ctx, self) during their PreStart. Domains and any other
// non-variant consumers construct a Factory directly via NewFactory.
//
// Factories are not stored in the global services registry; they are passed
// through a container scoped to the PreStart pass — see Pass.
package ddnsmetric

import constpkg "github.com/duakc/lightddns/constant"

const Namespace = constpkg.Project

const (
	SubsystemDomain     = "domain"
	SubsystemProvider   = "provider"
	SubsystemDatasource = "datasource"
	SubsystemService    = "service"
)

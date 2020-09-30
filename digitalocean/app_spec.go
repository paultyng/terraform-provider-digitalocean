package digitalocean

import (
	"github.com/digitalocean/godo"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func appSpecSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the app",
		},
		"region": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The slug for the DigitalOcean data center region hosting the app",
		},
		"domains": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"service": {
			Type:     schema.TypeList,
			Optional: true,
			MinItems: 1,
			Elem:     appSpecServicesSchema(),
		},
		"static_site": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     appSpecStaticSiteSchema(),
		},
		"worker": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     appSpecWorkerSchema(),
		},
		"database": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     appSpecDatabaseSchema(),
		},
	}
}

func appSpecGitSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo_clone_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"branch": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func appSpecGitHubSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"repo": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"branch": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"deploy_on_push": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}

func appSpecEnvSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"scope": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "RUN_AND_BUILD_TIME",
				ValidateFunc: validation.StringInSlice([]string{
					"UNSET",
					"RUN_TIME",
					"BUILD_TIME",
					"RUN_AND_BUILD_TIME",
				}, false),
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "GENERAL",
				ValidateFunc: validation.StringInSlice([]string{
					"GENERAL",
					"SECRET",
				}, false),
			},
		},
	}
}

func appSpecRouteSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Path specifies an route by HTTP path prefix. Paths must start with / and must be unique within the app.",
		},
	}
}

func appSpecHealthCheckSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"path": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Path is the route path used for the HTTP health check ping.",
		},
	}
}

func appSpecComponentBase() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the component",
		},
		"git": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: appSpecGitSourceSchema(),
			},
		},
		"github": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: appSpecGitHubSourceSchema(),
			},
		},
		"dockerfile_path": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"build_command": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"env": {
			Type:     schema.TypeSet,
			Optional: true,
			Elem:     appSpecEnvSchema(),
			Set:      schema.HashResource(appSpecEnvSchema()),
		},
		"routes": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: appSpecRouteSchema(),
			},
		},
		"source_dir": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"environment_slug": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func appSpecServicesSchema() *schema.Resource {
	serviceSchema := map[string]*schema.Schema{
		"run_command": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"http_port": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"instance_size_slug": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"instance_count": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"health_check": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: appSpecHealthCheckSchema(),
			},
		},
	}

	for k, v := range appSpecComponentBase() {
		serviceSchema[k] = v
	}

	return &schema.Resource{
		Schema: serviceSchema,
	}
}

func appSpecStaticSiteSchema() *schema.Resource {
	staticSiteSchema := map[string]*schema.Schema{
		"output_dir": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "An optional path to where the built assets will be located, relative to the build context. If not set, App Platform will automatically scan for these directory names: `_static`, `dist`, `public`.",
		},
		"index_document": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"error_document": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	for k, v := range appSpecComponentBase() {
		staticSiteSchema[k] = v
	}

	return &schema.Resource{
		Schema: staticSiteSchema,
	}
}

func appSpecWorkerSchema() *schema.Resource {
	workerSchema := map[string]*schema.Schema{
		"run_command": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"instance_size_slug": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"instance_count": {
			Type:     schema.TypeInt,
			Optional: true,
		},
	}

	for k, v := range appSpecComponentBase() {
		workerSchema[k] = v
	}

	return &schema.Resource{
		Schema: workerSchema,
	}
}

func appSpecDatabaseSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"UNSET",
					"MYSQL",
					"PG",
					"REDIS",
				}, false),
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"production": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"cluster_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"db_user": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func expandAppSpec(config []interface{}) *godo.AppSpec {
	if len(config) == 0 || config[0] == nil {
		return &godo.AppSpec{}
	}
	appSpecConfig := config[0].(map[string]interface{})

	appSpec := &godo.AppSpec{
		Name:        appSpecConfig["name"].(string),
		Region:      appSpecConfig["region"].(string),
		Domains:     expandAppDomainSpec(appSpecConfig["domains"].(*schema.Set).List()),
		Services:    expandAppSpecServices(appSpecConfig["service"].([]interface{})),
		StaticSites: expandAppSpecStaticSites(appSpecConfig["static_site"].([]interface{})),
		Workers:     expandAppSpecWorkers(appSpecConfig["worker"].([]interface{})),
		Databases:   expandAppSpecDatabases(appSpecConfig["database"].([]interface{})),
	}

	return appSpec
}

func flattenAppSpec(spec *godo.AppSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	if spec != nil {

		r := make(map[string]interface{})
		r["name"] = (*spec).Name
		r["region"] = (*spec).Region
		r["domains"] = flattenAppDomainSpec((*spec).Domains)

		if len((*spec).Services) > 0 {
			r["service"] = flattenAppSpecServices((*spec).Services)
		}

		if len((*spec).StaticSites) > 0 {
			r["static_site"] = flattenAppSpecStaticSites((*spec).StaticSites)
		}

		if len((*spec).Workers) > 0 {
			r["worker"] = flattenAppSpecWorkers((*spec).Workers)
		}

		if len((*spec).Databases) > 0 {
			r["database"] = flattenAppSpecDatabases((*spec).Databases)
		}

		result = append(result, r)
	}

	return result
}

func expandAppDomainSpec(config []interface{}) []*godo.AppDomainSpec {
	appDomains := make([]*godo.AppDomainSpec, 0, len(config))

	for _, rawDomain := range config {
		domain := &godo.AppDomainSpec{
			Domain: rawDomain.(string),
		}

		appDomains = append(appDomains, domain)
	}

	return appDomains
}

func flattenAppDomainSpec(spec []*godo.AppDomainSpec) *schema.Set {
	result := schema.NewSet(schema.HashString, []interface{}{})

	if spec != nil {
		for _, domain := range spec {
			result.Add(domain.Domain)
		}
	}

	return result
}

func expandAppGitHubSourceSpec(config []interface{}) *godo.GitHubSourceSpec {
	gitHubSourceConfig := config[0].(map[string]interface{})

	gitHubSource := &godo.GitHubSourceSpec{
		Repo:         gitHubSourceConfig["repo"].(string),
		Branch:       gitHubSourceConfig["branch"].(string),
		DeployOnPush: gitHubSourceConfig["deploy_on_push"].(bool),
	}

	return gitHubSource
}

func flattenAppGitHubSourceSpec(spec *godo.GitHubSourceSpec) []interface{} {
	result := make([]interface{}, 0)

	if spec != nil {

		r := make(map[string]interface{})
		r["repo"] = (*spec).Repo
		r["branch"] = (*spec).Branch
		r["deploy_on_push"] = (*spec).DeployOnPush

		result = append(result, r)
	}

	return result
}

func expandAppGitSourceSpec(config []interface{}) *godo.GitSourceSpec {
	gitSourceConfig := config[0].(map[string]interface{})

	gitSource := &godo.GitSourceSpec{
		Branch:       gitSourceConfig["branch"].(string),
		RepoCloneURL: gitSourceConfig["repo_clone_url"].(string),
	}

	return gitSource
}

func flattenAppGitSourceSpec(spec *godo.GitSourceSpec) []interface{} {
	result := make([]interface{}, 0)

	if spec != nil {

		r := make(map[string]interface{})
		r["branch"] = (*spec).Branch
		r["repo_clone_url"] = (*spec).RepoCloneURL

		result = append(result, r)
	}

	return result
}

func expandAppEnvs(config []interface{}) []*godo.AppVariableDefinition {
	appEnvs := make([]*godo.AppVariableDefinition, 0, len(config))

	for _, rawEnv := range config {
		env := rawEnv.(map[string]interface{})

		e := &godo.AppVariableDefinition{
			Value: env["value"].(string),
			Scope: godo.AppVariableScope(env["scope"].(string)),
			Key:   env["key"].(string),
			Type:  godo.AppVariableType(env["type"].(string)),
		}

		appEnvs = append(appEnvs, e)
	}

	return appEnvs
}

func flattenAppEnvs(appEnvs []*godo.AppVariableDefinition) *schema.Set {
	result := schema.NewSet(schema.HashResource(appSpecEnvSchema()), []interface{}{})

	if appEnvs != nil {
		for _, env := range appEnvs {
			r := make(map[string]interface{})
			r["value"] = env.Value
			r["scope"] = string(env.Scope)
			r["key"] = env.Key
			r["type"] = string(env.Type)

			result.Add(r)
		}
	}

	return result
}

func expandAppHealthCheck(config []interface{}) *godo.AppServiceSpecHealthCheck {
	healthCheckConfig := config[0].(map[string]interface{})

	healthCheck := &godo.AppServiceSpecHealthCheck{
		Path: healthCheckConfig["path"].(string),
	}

	return healthCheck
}

func flattenAppHealthCheck(check *godo.AppServiceSpecHealthCheck) []interface{} {
	result := make([]interface{}, 0)

	if check != nil {

		r := make(map[string]interface{})
		r["path"] = check.Path

		result = append(result, r)
	}

	return result
}

func expandAppRoutes(config []interface{}) []*godo.AppRouteSpec {
	appRoutes := make([]*godo.AppRouteSpec, 0, len(config))

	for _, rawRoute := range config {
		route := rawRoute.(map[string]interface{})

		r := &godo.AppRouteSpec{
			Path: route["path"].(string),
		}

		appRoutes = append(appRoutes, r)
	}

	return appRoutes
}

func flattenAppRoutes(routes []*godo.AppRouteSpec) []interface{} {
	result := make([]interface{}, 0)

	if routes != nil {
		for _, route := range routes {
			r := make(map[string]interface{})
			r["path"] = route.Path

			result = append(result, r)
		}
	}

	return result
}

func expandAppSpecServices(config []interface{}) []*godo.AppServiceSpec {
	appServices := make([]*godo.AppServiceSpec, 0, len(config))

	for _, rawService := range config {
		service := rawService.(map[string]interface{})

		s := &godo.AppServiceSpec{
			Name:             service["name"].(string),
			RunCommand:       service["run_command"].(string),
			BuildCommand:     service["build_command"].(string),
			HTTPPort:         int64(service["http_port"].(int)),
			DockerfilePath:   service["dockerfile_path"].(string),
			Envs:             expandAppEnvs(service["env"].(*schema.Set).List()),
			InstanceSizeSlug: service["instance_size_slug"].(string),
			InstanceCount:    int64(service["instance_count"].(int)),
			SourceDir:        service["source_dir"].(string),
			EnvironmentSlug:  service["environment_slug"].(string),
		}

		github := service["github"].([]interface{})
		if len(github) > 0 {
			s.GitHub = expandAppGitHubSourceSpec(github)
		}

		git := service["git"].([]interface{})
		if len(git) > 0 {
			s.Git = expandAppGitSourceSpec(git)
		}

		routes := service["routes"].([]interface{})
		if len(routes) > 0 {
			s.Routes = expandAppRoutes(routes)
		}

		checks := service["health_check"].([]interface{})
		if len(checks) > 0 {
			s.HealthCheck = expandAppHealthCheck(checks)
		}

		appServices = append(appServices, s)
	}

	return appServices
}

func flattenAppSpecServices(services []*godo.AppServiceSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, len(services))

	for i, s := range services {
		r := make(map[string]interface{})

		r["name"] = s.Name
		r["run_command"] = s.RunCommand
		r["build_command"] = s.BuildCommand
		r["github"] = flattenAppGitHubSourceSpec(s.GitHub)
		r["git"] = flattenAppGitSourceSpec(s.Git)
		r["http_port"] = int(s.HTTPPort)
		r["routes"] = flattenAppRoutes(s.Routes)
		r["dockerfile_path"] = s.DockerfilePath
		r["env"] = flattenAppEnvs(s.Envs)
		r["health_check"] = flattenAppHealthCheck(s.HealthCheck)
		r["instance_size_slug"] = s.InstanceSizeSlug
		r["instance_count"] = int(s.InstanceCount)
		r["source_dir"] = s.SourceDir
		r["environment_slug"] = s.EnvironmentSlug

		result[i] = r
	}

	return result
}

func expandAppSpecStaticSites(config []interface{}) []*godo.AppStaticSiteSpec {
	appSites := make([]*godo.AppStaticSiteSpec, 0, len(config))

	for _, rawSite := range config {
		site := rawSite.(map[string]interface{})

		s := &godo.AppStaticSiteSpec{
			Name:            site["name"].(string),
			BuildCommand:    site["build_command"].(string),
			DockerfilePath:  site["dockerfile_path"].(string),
			Envs:            expandAppEnvs(site["env"].(*schema.Set).List()),
			SourceDir:       site["source_dir"].(string),
			OutputDir:       site["output_dir"].(string),
			IndexDocument:   site["index_document"].(string),
			ErrorDocument:   site["error_document"].(string),
			EnvironmentSlug: site["environment_slug"].(string),
		}

		github := site["github"].([]interface{})
		if len(github) > 0 {
			s.GitHub = expandAppGitHubSourceSpec(github)
		}

		git := site["git"].([]interface{})
		if len(git) > 0 {
			s.Git = expandAppGitSourceSpec(git)
		}

		routes := site["routes"].([]interface{})
		if len(routes) > 0 {
			s.Routes = expandAppRoutes(routes)
		}

		appSites = append(appSites, s)
	}

	return appSites
}

func flattenAppSpecStaticSites(sites []*godo.AppStaticSiteSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, len(sites))

	for i, s := range sites {
		r := make(map[string]interface{})

		r["name"] = s.Name
		r["build_command"] = s.BuildCommand
		r["github"] = flattenAppGitHubSourceSpec(s.GitHub)
		r["git"] = flattenAppGitSourceSpec(s.Git)
		r["routes"] = flattenAppRoutes(s.Routes)
		r["dockerfile_path"] = s.DockerfilePath
		r["env"] = flattenAppEnvs(s.Envs)
		r["source_dir"] = s.SourceDir
		r["output_dir"] = s.OutputDir
		r["index_document"] = s.IndexDocument
		r["error_document"] = s.ErrorDocument
		r["environment_slug"] = s.EnvironmentSlug

		result[i] = r
	}

	return result
}

func expandAppSpecWorkers(config []interface{}) []*godo.AppWorkerSpec {
	appWorkers := make([]*godo.AppWorkerSpec, 0, len(config))

	for _, rawWorker := range config {
		worker := rawWorker.(map[string]interface{})

		s := &godo.AppWorkerSpec{
			Name:             worker["name"].(string),
			RunCommand:       worker["run_command"].(string),
			BuildCommand:     worker["build_command"].(string),
			DockerfilePath:   worker["dockerfile_path"].(string),
			Envs:             expandAppEnvs(worker["env"].(*schema.Set).List()),
			InstanceSizeSlug: worker["instance_size_slug"].(string),
			InstanceCount:    int64(worker["instance_count"].(int)),
			SourceDir:        worker["source_dir"].(string),
			EnvironmentSlug:  worker["environment_slug"].(string),
		}

		github := worker["github"].([]interface{})
		if len(github) > 0 {
			s.GitHub = expandAppGitHubSourceSpec(github)
		}

		git := worker["git"].([]interface{})
		if len(git) > 0 {
			s.Git = expandAppGitSourceSpec(git)
		}

		appWorkers = append(appWorkers, s)
	}

	return appWorkers
}

func flattenAppSpecWorkers(workers []*godo.AppWorkerSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, len(workers))

	for i, w := range workers {
		r := make(map[string]interface{})

		r["name"] = w.Name
		r["run_command"] = w.RunCommand
		r["build_command"] = w.BuildCommand
		r["github"] = flattenAppGitHubSourceSpec(w.GitHub)
		r["git"] = flattenAppGitSourceSpec(w.Git)
		r["dockerfile_path"] = w.DockerfilePath
		r["env"] = flattenAppEnvs(w.Envs)
		r["instance_size_slug"] = w.InstanceSizeSlug
		r["instance_count"] = int(w.InstanceCount)
		r["source_dir"] = w.SourceDir
		r["environment_slug"] = w.EnvironmentSlug

		result[i] = r
	}

	return result
}

func expandAppSpecDatabases(config []interface{}) []*godo.AppDatabaseSpec {
	appDatabases := make([]*godo.AppDatabaseSpec, 0, len(config))

	for _, rawDatabase := range config {
		db := rawDatabase.(map[string]interface{})

		s := &godo.AppDatabaseSpec{
			Name:        db["name"].(string),
			Engine:      godo.AppDatabaseSpecEngine(db["engine"].(string)),
			Version:     db["version"].(string),
			Production:  db["production"].(bool),
			ClusterName: db["cluster_name"].(string),
			DBName:      db["db_name"].(string),
			DBUser:      db["db_user"].(string),
		}

		appDatabases = append(appDatabases, s)
	}

	return appDatabases
}

func flattenAppSpecDatabases(databases []*godo.AppDatabaseSpec) []map[string]interface{} {
	result := make([]map[string]interface{}, len(databases))

	for i, db := range databases {
		r := make(map[string]interface{})

		r["name"] = db.Name
		r["engine"] = db.Engine
		r["version"] = db.Version
		r["production"] = db.Production
		r["cluster_name"] = db.ClusterName
		r["db_name"] = db.DBName
		r["db_user"] = db.DBUser

		result[i] = r
	}

	return result
}

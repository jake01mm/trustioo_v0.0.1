// Package deployment 提供部署工具和脚本
package deployment

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// DeploymentConfig 部署配置
type DeploymentConfig struct {
	Environment string            `json:"environment"`
	Namespace   string            `json:"namespace"`
	ImageTag    string            `json:"image_tag"`
	Replicas    int               `json:"replicas"`
	Resources   ResourceConfig    `json:"resources"`
	EnvVars     map[string]string `json:"env_vars"`
	Secrets     map[string]string `json:"secrets"`
	HealthCheck HealthCheckConfig `json:"health_check"`
}

// ResourceConfig 资源配置
type ResourceConfig struct {
	CPU    ResourceSpec `json:"cpu"`
	Memory ResourceSpec `json:"memory"`
}

// ResourceSpec 资源规格
type ResourceSpec struct {
	Requests string `json:"requests"`
	Limits   string `json:"limits"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Path                string `json:"path"`
	InitialDelaySeconds int    `json:"initial_delay_seconds"`
	PeriodSeconds       int    `json:"period_seconds"`
	TimeoutSeconds      int    `json:"timeout_seconds"`
	FailureThreshold    int    `json:"failure_threshold"`
}

// Deployer 部署器
type Deployer struct {
	logger *logrus.Logger
	config *DeploymentConfig
}

// NewDeployer 创建部署器
func NewDeployer(config *DeploymentConfig, logger *logrus.Logger) *Deployer {
	return &Deployer{
		logger: logger,
		config: config,
	}
}

// Deploy 执行部署
func (d *Deployer) Deploy(ctx context.Context) error {
	d.logger.WithFields(logrus.Fields{
		"environment": d.config.Environment,
		"namespace":   d.config.Namespace,
		"image_tag":   d.config.ImageTag,
	}).Info("Starting deployment")

	// 检查前置条件
	if err := d.checkPrerequisites(); err != nil {
		return fmt.Errorf("prerequisites check failed: %w", err)
	}

	// 构建Docker镜像
	if err := d.buildImage(ctx); err != nil {
		return fmt.Errorf("image build failed: %w", err)
	}

	// 推送镜像
	if err := d.pushImage(ctx); err != nil {
		return fmt.Errorf("image push failed: %w", err)
	}

	// 部署到Kubernetes
	if err := d.deployToKubernetes(ctx); err != nil {
		return fmt.Errorf("kubernetes deployment failed: %w", err)
	}

	// 验证部署
	if err := d.verifyDeployment(ctx); err != nil {
		return fmt.Errorf("deployment verification failed: %w", err)
	}

	d.logger.Info("Deployment completed successfully")
	return nil
}

// checkPrerequisites 检查前置条件
func (d *Deployer) checkPrerequisites() error {
	// 检查Docker
	if err := d.runCommand("docker", "--version"); err != nil {
		return fmt.Errorf("docker not available: %w", err)
	}

	// 检查kubectl
	if err := d.runCommand("kubectl", "version", "--client"); err != nil {
		return fmt.Errorf("kubectl not available: %w", err)
	}

	// 检查Kubernetes连接
	if err := d.runCommand("kubectl", "cluster-info"); err != nil {
		return fmt.Errorf("kubernetes cluster not accessible: %w", err)
	}

	return nil
}

// buildImage 构建Docker镜像
func (d *Deployer) buildImage(ctx context.Context) error {
	d.logger.Info("Building Docker image")

	imageName := fmt.Sprintf("trusioo/api:%s", d.config.ImageTag)

	cmd := exec.CommandContext(ctx, "docker", "build",
		"-f", "docker/Dockerfile.prod",
		"-t", imageName,
		".")

	cmd.Dir = d.getProjectRoot()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// pushImage 推送Docker镜像
func (d *Deployer) pushImage(ctx context.Context) error {
	d.logger.Info("Pushing Docker image")

	imageName := fmt.Sprintf("trusioo/api:%s", d.config.ImageTag)

	cmd := exec.CommandContext(ctx, "docker", "push", imageName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// deployToKubernetes 部署到Kubernetes
func (d *Deployer) deployToKubernetes(ctx context.Context) error {
	d.logger.Info("Deploying to Kubernetes")

	// 创建命名空间（如果不存在）
	if err := d.createNamespace(ctx); err != nil {
		return err
	}

	// 应用ConfigMap和Secret
	if err := d.applyConfigs(ctx); err != nil {
		return err
	}

	// 应用部署配置
	manifestPath := filepath.Join(d.getProjectRoot(), "k8s", "deployment.yaml")
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", manifestPath, "-n", d.config.Namespace)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// createNamespace 创建命名空间
func (d *Deployer) createNamespace(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "kubectl", "create", "namespace", d.config.Namespace, "--dry-run=client", "-o", "yaml")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(string(output))
	return cmd.Run()
}

// applyConfigs 应用配置
func (d *Deployer) applyConfigs(ctx context.Context) error {
	// 应用ConfigMap
	configMapData := d.generateConfigMap()
	if err := d.applyManifest(ctx, configMapData); err != nil {
		return fmt.Errorf("failed to apply configmap: %w", err)
	}

	// 应用Secret
	secretData := d.generateSecret()
	if err := d.applyManifest(ctx, secretData); err != nil {
		return fmt.Errorf("failed to apply secret: %w", err)
	}

	return nil
}

// generateConfigMap 生成ConfigMap配置
func (d *Deployer) generateConfigMap() string {
	return fmt.Sprintf(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: trusioo-api-config
  namespace: %s
data:
  GIN_MODE: release
  LOG_LEVEL: %s
`, d.config.Namespace, d.getLogLevel())
}

// generateSecret 生成Secret配置
func (d *Deployer) generateSecret() string {
	return fmt.Sprintf(`
apiVersion: v1
kind: Secret
metadata:
  name: trusioo-api-secrets
  namespace: %s
type: Opaque
stringData:
  JWT_SECRET: "%s"
  PASSWORD_ENCRYPTION_KEY: "%s"
`, d.config.Namespace, d.getSecret("JWT_SECRET"), d.getSecret("PASSWORD_ENCRYPTION_KEY"))
}

// applyManifest 应用清单
func (d *Deployer) applyManifest(ctx context.Context, manifest string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(manifest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// verifyDeployment 验证部署
func (d *Deployer) verifyDeployment(ctx context.Context) error {
	d.logger.Info("Verifying deployment")

	// 等待部署完成
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("deployment verification timed out")
		case <-ticker.C:
			if ready, err := d.checkDeploymentReady(ctx); err != nil {
				d.logger.WithError(err).Warn("Error checking deployment status")
			} else if ready {
				d.logger.Info("Deployment is ready")
				return d.verifyHealthCheck(ctx)
			}
		}
	}
}

// checkDeploymentReady 检查部署是否就绪
func (d *Deployer) checkDeploymentReady(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "get", "deployment", "trusioo-api",
		"-n", d.config.Namespace, "-o", "jsonpath={.status.readyReplicas}")

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	readyReplicas := strings.TrimSpace(string(output))
	expectedReplicas := fmt.Sprintf("%d", d.config.Replicas)

	return readyReplicas == expectedReplicas, nil
}

// verifyHealthCheck 验证健康检查
func (d *Deployer) verifyHealthCheck(ctx context.Context) error {
	d.logger.Info("Verifying health check")

	// 获取服务端点
	cmd := exec.CommandContext(ctx, "kubectl", "get", "service", "trusioo-api-service",
		"-n", d.config.Namespace, "-o", "jsonpath={.spec.clusterIP}")

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	serviceIP := strings.TrimSpace(string(output))
	healthURL := fmt.Sprintf("http://%s%s", serviceIP, d.config.HealthCheck.Path)

	// 使用kubectl exec进行健康检查
	podName, err := d.getFirstPodName(ctx)
	if err != nil {
		return err
	}

	cmd = exec.CommandContext(ctx, "kubectl", "exec", podName, "-n", d.config.Namespace,
		"--", "wget", "--spider", "-q", healthURL)

	return cmd.Run()
}

// getFirstPodName 获取第一个Pod名称
func (d *Deployer) getFirstPodName(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "kubectl", "get", "pods",
		"-l", "app=trusioo-api", "-n", d.config.Namespace,
		"-o", "jsonpath={.items[0].metadata.name}")

	output, err := cmd.Output()
	return strings.TrimSpace(string(output)), err
}

// runCommand 运行命令
func (d *Deployer) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getProjectRoot 获取项目根目录
func (d *Deployer) getProjectRoot() string {
	// 简化实现，实际应该动态查找
	return "."
}

// getLogLevel 获取日志级别
func (d *Deployer) getLogLevel() string {
	switch d.config.Environment {
	case "production":
		return "warn"
	case "staging":
		return "info"
	default:
		return "debug"
	}
}

// getSecret 获取密钥
func (d *Deployer) getSecret(key string) string {
	if value, exists := d.config.Secrets[key]; exists {
		return value
	}
	return os.Getenv(key)
}

// Rollback 回滚部署
func (d *Deployer) Rollback(ctx context.Context, revision string) error {
	d.logger.WithFields(logrus.Fields{
		"namespace": d.config.Namespace,
		"revision":  revision,
	}).Info("Rolling back deployment")

	args := []string{"rollout", "undo", "deployment/trusioo-api", "-n", d.config.Namespace}
	if revision != "" {
		args = append(args, "--to-revision="+revision)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Scale 扩缩容
func (d *Deployer) Scale(ctx context.Context, replicas int) error {
	d.logger.WithFields(logrus.Fields{
		"namespace": d.config.Namespace,
		"replicas":  replicas,
	}).Info("Scaling deployment")

	cmd := exec.CommandContext(ctx, "kubectl", "scale", "deployment", "trusioo-api",
		"-n", d.config.Namespace, "--replicas="+fmt.Sprintf("%d", replicas))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetStatus 获取部署状态
func (d *Deployer) GetStatus(ctx context.Context) (map[string]any, error) {
	// 获取部署状态
	cmd := exec.CommandContext(ctx, "kubectl", "get", "deployment", "trusioo-api",
		"-n", d.config.Namespace, "-o", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 简化实现，实际应该解析JSON
	status := map[string]any{
		"deployment_output": string(output),
		"timestamp":         time.Now(),
		"namespace":         d.config.Namespace,
	}

	return status, nil
}

// DefaultDeploymentConfig 返回默认部署配置
func DefaultDeploymentConfig(environment string) *DeploymentConfig {
	config := &DeploymentConfig{
		Environment: environment,
		Namespace:   "trusioo-api",
		ImageTag:    "latest",
		Replicas:    3,
		Resources: ResourceConfig{
			CPU: ResourceSpec{
				Requests: "250m",
				Limits:   "500m",
			},
			Memory: ResourceSpec{
				Requests: "256Mi",
				Limits:   "512Mi",
			},
		},
		HealthCheck: HealthCheckConfig{
			Path:                "/health",
			InitialDelaySeconds: 30,
			PeriodSeconds:       10,
			TimeoutSeconds:      3,
			FailureThreshold:    3,
		},
	}

	// 根据环境调整配置
	switch environment {
	case "production":
		config.Replicas = 5
		config.Resources.CPU.Limits = "1000m"
		config.Resources.Memory.Limits = "1Gi"
	case "staging":
		config.Replicas = 2
	case "development":
		config.Replicas = 1
		config.Resources.CPU.Requests = "100m"
		config.Resources.Memory.Requests = "128Mi"
	}

	return config
}

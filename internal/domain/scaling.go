// internal/domain/scaling.go
package domain

import "time"

// ==================== HPA (Horizontal Pod Autoscaler) ====================

// HPAName định nghĩa kiểu cho tên HPA
type HPAName string

// HPAMetricType định nghĩa kiểu cho loại Metric của HPA
type HPAMetricType string

const (
	MetricTypeResource HPAMetricType = "Resource"
	MetricTypePods     HPAMetricType = "Pods"
	MetricTypeObject   HPAMetricType = "Object"
	MetricTypeExternal HPAMetricType = "External"
)

// HPAStatusType định nghĩa kiểu cho trạng thái HPA (ví dụ: ScalingActive, AbleToScale)
type HPAStatusType string

// HPAMetric mô hình hóa một Metric được HPA theo dõi
type HPAMetric struct {
	Type         HPAMetricType `json:"type"`
	ResourceName string        `json:"resource_name,omitempty"` // Tên Resource (vd: cpu, memory)
	TargetType   string        `json:"target_type,omitempty"`   // Loại Target (vd: AverageUtilization)
	TargetValue  string        `json:"target_value"`            // Giá trị Target (vd: 80% hoặc 500m)
	CurrentValue string        `json:"current_value"`           // Giá trị hiện tại
}

// HPA mô hình hóa đối tượng Horizontal Pod Autoscaler
type HPA struct {
	Name            HPAName           `json:"name"`
	Namespace       Namespace         `json:"namespace"`
	TargetKind      string            `json:"target_kind"`      // Loại tài nguyên mục tiêu (vd: Deployment)
	TargetName      string            `json:"target_name"`      // Tên tài nguyên mục tiêu
	MinReplicas     int32             `json:"min_replicas"`     // Số lượng Pod tối thiểu
	MaxReplicas     int32             `json:"max_replicas"`     // Số lượng Pod tối đa
	CurrentReplicas int32             `json:"current_replicas"` // Số lượng Pod hiện tại
	DesiredReplicas int32             `json:"desired_replicas"` // Số lượng Pod mong muốn
	Metrics         []HPAMetric       `json:"metrics"`
	Conditions      []HPACondition    `json:"conditions"`
	Labels          map[string]string `json:"labels,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
}

// HPACondition mô hình hóa trạng thái (Condition) của HPA
type HPACondition struct {
	Type           HPAStatusType `json:"type"`
	Status         string        `json:"status"` // True, False, Unknown
	Reason         string        `json:"reason,omitempty"`
	Message        string        `json:"message,omitempty"`
	LastTransition time.Time     `json:"last_transition"`
}

// ==================== LimitRange & ResourceQuota (Cần thiết cho Resources) ====================

// ResourceQuotaName định nghĩa kiểu cho tên ResourceQuota
type ResourceQuotaName string

// ResourceQuota mô hình hóa đối tượng ResourceQuota
type ResourceQuota struct {
	Name      ResourceQuotaName `json:"name"`
	Namespace Namespace         `json:"namespace"`
	HardLimit map[string]string `json:"hard_limit"` // Giới hạn tổng cộng
	Used      map[string]string `json:"used"`       // Đã sử dụng
	Status    string            `json:"status"`     // Tóm tắt trạng thái (ví dụ: có vượt quá giới hạn không)
	CreatedAt time.Time         `json:"created_at"`
}

// LimitRangeName định nghĩa kiểu cho tên LimitRange
type LimitRangeName string

// LimitRangeItem mô hình hóa một mục giới hạn trong LimitRange
type LimitRangeItem struct {
	Type           string            `json:"type"` // Pod, Container, PersistentVolumeClaim
	Max            map[string]string `json:"max"`
	Min            map[string]string `json:"min"`
	Default        map[string]string `json:"default"`
	DefaultRequest map[string]string `json:"default_request"`
}

// LimitRange mô hình hóa đối tượng LimitRange
type LimitRange struct {
	Name      LimitRangeName   `json:"name"`
	Namespace Namespace        `json:"namespace"`
	Limits    []LimitRangeItem `json:"limits"`
	CreatedAt time.Time        `json:"created_at"`
}

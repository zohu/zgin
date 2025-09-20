package zauth

import (
	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
	"github.com/zohu/zgin/zlog"

	"context"
	"fmt"
)

func GoogleCaptcha(ctx context.Context, action, token, pid, secret string) (float64, error) {
	client, err := recaptcha.NewClient(ctx)
	if err != nil {
		zlog.Errorf("Captcha 创建验证器失败 %s %v\n %s", pid, err, token)
		return 1, fmt.Errorf("captcha failed")
	}
	defer client.Close()

	// 构建评估请求。
	response, err := client.CreateAssessment(
		ctx,
		&recaptchapb.CreateAssessmentRequest{
			Assessment: &recaptchapb.Assessment{
				Event: &recaptchapb.Event{
					Token:   token,
					SiteKey: secret,
				},
			},
			Parent: fmt.Sprintf("projects/%s", pid),
		},
	)
	if err != nil {
		zlog.Warnf("Captcha 校验失败: %v", err.Error())
		return 1, fmt.Errorf("captcha failed")
	}
	// 检查令牌是否有效。
	if !response.TokenProperties.Valid {
		zlog.Warnf("Captcha 非法Token: %v",
			response.TokenProperties.InvalidReason)
		return 0, fmt.Errorf("captcha token failed")
	}
	// 检查是否执行了预期操作。
	if response.TokenProperties.Action != action {
		zlog.Warnf("Captcha Action 不符")
		return 0, fmt.Errorf("captcha action failed")
	}
	return float64(response.RiskAnalysis.Score), nil
}

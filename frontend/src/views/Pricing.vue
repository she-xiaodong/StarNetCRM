<template>
  <div class="pricing-page">
    <div class="page-hero">
      <h1>选择适合你的方案</h1>
      <p>14天免费试用，无需绑定信用卡，随时取消</p>
    </div>

    <div class="plan-grid">
      <div class="plan-card" v-for="plan in plans" :key="plan.name" :class="{ featured: plan.featured }">
        <div v-if="plan.featured" class="popular-ribbon">最受欢迎</div>
        <div class="plan-header">
          <h3>{{ plan.name }}</h3>
          <p>{{ plan.desc }}</p>
        </div>
        <div class="plan-price">
          <span class="symbol">¥</span>
          <span class="num">{{ plan.price }}</span>
          <span class="unit">/人/月</span>
        </div>
        <div class="plan-body">
          <div class="feature-title">包含功能</div>
          <ul>
            <li v-for="f in plan.features" :key="f">
              <check-circle-outlined /> {{ f }}
            </li>
          </ul>
        </div>
        <div class="plan-footer">
          <a-button :type="plan.featured ? 'primary' : 'default'" size="large" block @click="handleSelect(plan)">
            {{ plan.cta }}
          </a-button>
        </div>
      </div>
    </div>

    <div class="faq-section">
      <h2>常见问题</h2>
      <a-collapse :bordered="false" accordion>
        <a-collapse-panel v-for="faq in faqs" :key="faq.q" :header="faq.q">
          <p>{{ faq.a }}</p>
        </a-collapse-panel>
      </a-collapse>
    </div>
  </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
import { CheckCircleOutlined } from '@ant-design/icons-vue'

const plans = [
  {
    name: '免费版',
    price: 0,
    desc: '个人体验和小团队入门',
    features: [
      '100个人脉上限',
      '基础关系图谱',
      '标签管理',
      '联系人导入/导出',
      '社区支持',
    ],
    cta: '免费开始',
    featured: false,
  },
  {
    name: '专业版',
    price: 29,
    desc: '适合销售、BD和猎头团队',
    features: [
      '无限人脉记录',
      '完整关系图谱',
      '六度路径搜索',
      '引荐管理工作流',
      '数据分析报表',
      'API 接口',
      '邮件 + 在线支持',
      '5个团队成员',
    ],
    cta: '免费试用14天',
    featured: true,
  },
  {
    name: '企业版',
    price: 59,
    desc: '适合大型组织和深度用户',
    features: [
      '专业版全部功能',
      '无限制团队成员',
      '多租户管理',
      '自定义字段和流程',
      'SSO单点登录',
      '私有化部署选项',
      '专属客户成功经理',
      '7×24小时支持',
    ],
    cta: '联系我们',
    featured: false,
  },
]

const faqs = [
  { q: '免费版有什么限制？', a: '免费版最多可录入100个联系人，可以使用基础的关系图谱和标签管理功能。适合个人用户和小团队体验使用。' },
  { q: '如何从免费版升级？', a: '在工作台设置中点击升级按钮，选择适合的方案即可。升级后所有数据无缝迁移，不会丢失任何信息。' },
  { q: '支持哪些支付方式？', a: '目前支持支付宝和微信支付，企业用户支持对公转账和发票申请。' },
  { q: '可以随时取消订阅吗？', a: '可以。月付计划随时取消，已付费用不予退还。年付计划支持7天内无理由退款。' },
  { q: '数据安全如何保障？', a: '所有数据传输使用TLS加密，数据存储AES-256加密。支持私有化部署，数据完全由您掌控。' },
  { q: '"团队成员"如何计算？', a: '按系统中激活的账号数量计算，已禁用/删除的账号不计入统计。' },
]

function handleSelect(plan: any) {
  if (plan.price === 0) {
    message.info('注册后即可使用免费版')
  } else if (plan.price === 59) {
    message.info('请联系我们的销售团队获取企业版方案')
  } else {
    message.info('注册后可选择专业版试用')
  }
}
</script>

<style scoped lang="less">
@primary: #2B5FD7;
@primary-light: #4A7AE8;
@text: #1A1A1A;
@text-secondary: #666;

.page-hero {
  text-align: center; padding: 60px 24px 40px;
  h1 { font-size: 36px; font-weight: 700; color: @text; margin: 0 0 12px; }
  p { font-size: 16px; color: @text-secondary; margin: 0; }
}

.plan-grid {
  max-width: 1100px; margin: 0 auto; padding: 0 24px 60px;
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px;
}

.plan-card {
  background: #fff; border-radius: 16px; padding: 36px 28px;
  border: 2px solid #eee; position: relative; transition: all 0.3s;
  display: flex; flex-direction: column;
  &:hover { transform: translateY(-4px); box-shadow: 0 12px 40px rgba(0,0,0,0.08); }
  &.featured {
    border-color: @primary; box-shadow: 0 8px 36px rgba(43,95,215,0.12);
    .plan-price .num { color: @primary; }
  }
}

.popular-ribbon {
  position: absolute; top: -12px; left: 50%; transform: translateX(-50%);
  background: linear-gradient(135deg, @primary, @primary-light);
  color: #fff; padding: 4px 24px; border-radius: 12px; font-size: 12px; font-weight: 600;
}

.plan-header {
  text-align: center; margin-bottom: 20px;
  h3 { font-size: 22px; font-weight: 700; margin: 0 0 6px; }
  p { font-size: 13px; color: @text-secondary; margin: 0; }
}

.plan-price {
  text-align: center; padding: 16px 0; border-top: 1px solid #f0f0f0; border-bottom: 1px solid #f0f0f0;
  margin-bottom: 20px;
  .symbol { font-size: 20px; font-weight: 600; vertical-align: top; }
  .num { font-size: 48px; font-weight: 800; }
  .unit { font-size: 14px; color: @text-secondary; }
}

.plan-body {
  flex: 1;
  .feature-title { font-size: 13px; font-weight: 600; color: @text-secondary; margin-bottom: 8px; text-transform: uppercase; }
  ul { list-style: none; padding: 0; margin: 0;
    li { padding: 7px 0; font-size: 14px; color: @text-secondary; display: flex; align-items: center; gap: 8px;
      .anticon { color: #00C9A7; font-size: 16px; }
    }
  }
}

.plan-footer { margin-top: 24px; }

.faq-section {
  max-width: 720px; margin: 0 auto; padding: 0 24px 60px;
  h2 { text-align: center; font-size: 28px; font-weight: 700; margin-bottom: 32px; }
  p { color: @text-secondary; line-height: 1.7; }
}

@media (max-width: 768px) {
  .plan-grid { grid-template-columns: 1fr; }
}
</style>

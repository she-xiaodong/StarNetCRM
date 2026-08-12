<template>
  <div class="landing">
    <!-- ─── 顶部导航 ─── -->
    <header class="nav-header" :class="{ scrolled: isScrolled }">
      <div class="nav-inner">
        <div class="nav-brand" @click="scrollToTop">
          <span class="brand-icon">✦</span>
          <span class="brand-text">星络客</span>
          <span class="brand-sub">StarNet CRM</span>
        </div>
        <nav class="nav-links">
          <a @click="scrollTo('features')">产品功能</a>
          <a @click="scrollTo('pricing')">定价方案</a>
          <a @click="scrollTo('cases')">应用案例</a>
          <a href="/dashboard" v-if="auth.isLoggedIn">进入工作台</a>
          <template v-else>
            <a-button type="text" @click="showLogin = true">登录</a-button>
            <a-button type="primary" @click="showRegister = true">免费注册</a-button>
          </template>
        </nav>
      </div>
    </header>

    <!-- ─── HERO 主视觉区 ─── -->
    <section class="hero">
      <div class="hero-bg"></div>
      <div class="hero-content">
        <div class="hero-badge">基于六度人脉理论的关系型 CRM</div>
        <h1 class="hero-title">
          你的每个人脉，<br />都值得被<span class="highlight">连接</span>
        </h1>
        <p class="hero-desc">
          星络客帮你发现隐藏的人脉价值，通过智能图谱、路径搜索和引荐管理，<br />
          让六度人脉理论在商业场景中真正落地。
        </p>
        <div class="hero-actions">
          <a-button type="primary" size="large" class="btn-cta" @click="showRegister = true">
            免费开始使用
          </a-button>
          <a-button size="large" class="btn-demo" @click="scrollTo('features')">
            了解产品
          </a-button>
        </div>
        <div class="hero-stats">
          <div class="stat-item">
            <span class="stat-num">3,500+</span>
            <span class="stat-label">企业客户</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <span class="stat-num">120万+</span>
            <span class="stat-label">人脉连接</span>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <span class="stat-num">99.9%</span>
            <span class="stat-label">服务可用性</span>
          </div>
        </div>
      </div>
    </section>

    <!-- ─── 产品功能 ─── -->
    <section id="features" class="features">
      <div class="section-head">
        <h2>核心功能</h2>
        <p>覆盖人脉管理的全生命周期，让每一段关系都创造价值</p>
      </div>
      <div class="feature-grid">
        <div class="feature-card" v-for="f in features" :key="f.title">
          <div class="feature-icon" :style="{ background: f.bg }">
            <component :is="f.icon" />
          </div>
          <h3>{{ f.title }}</h3>
          <p>{{ f.desc }}</p>
        </div>
      </div>
    </section>

    <!-- ─── 关系图谱亮点 ─── -->
    <section class="showcase">
      <div class="section-head light">
        <h2>可视化的关系图谱</h2>
        <p>直观展示人脉网络，一键发现关键连接者</p>
      </div>
      <div class="showcase-grid">
        <div class="showcase-item">
          <div class="showcase-img img-graph"></div>
          <h4>1度关系图谱</h4>
          <p>以你为中心，智能展示所有直接联系人及其关系强度</p>
        </div>
        <div class="showcase-item">
          <div class="showcase-img img-path"></div>
          <h4>六度路径搜索</h4>
          <p>输入起点和终点，系统自动计算最优人脉路径</p>
        </div>
        <div class="showcase-item">
          <div class="showcase-img img-stats"></div>
          <h4>人脉数据分析</h4>
          <p>行业分布、关系类型、活跃度等多维度数据分析</p>
        </div>
      </div>
    </section>

    <!-- ─── 定价方案 ─── -->
    <section id="pricing" class="pricing">
      <div class="section-head">
        <h2>灵活的定价方案</h2>
        <p>按需选择，随时升级，满足不同规模团队的需求</p>
      </div>
      <div class="pricing-grid">
        <div
          class="pricing-card"
          v-for="plan in plans"
          :key="plan.name"
          :class="{ featured: plan.featured }"
        >
          <div v-if="plan.featured" class="popular-badge">最受欢迎</div>
          <h3>{{ plan.name }}</h3>
          <div class="price">
            <span class="currency">¥</span>
            <span class="amount">{{ plan.price }}</span>
            <span class="period">/人/月</span>
          </div>
          <p class="plan-desc">{{ plan.desc }}</p>
          <ul class="plan-features">
            <li v-for="f in plan.features" :key="f">
              <check-outlined /> {{ f }}
            </li>
          </ul>
          <a-button
            :type="plan.featured ? 'primary' : 'default'"
            size="large"
            block
            @click="showRegister = true"
          >
            {{ plan.cta }}
          </a-button>
        </div>
      </div>
    </section>

    <!-- ─── 应用案例 ─── -->
    <section id="cases" class="cases">
      <div class="section-head light">
        <h2>他们都在使用星络客</h2>
        <p>覆盖30+行业，服务3500+企业客户</p>
      </div>
      <div class="case-logos">
        <span>金融</span><span>科技</span><span>教育</span><span>医疗</span>
        <span>制造</span><span>零售</span><span>咨询</span><span>房地产</span>
      </div>
    </section>

    <!-- ─── 页脚 ─── -->
    <footer class="footer">
      <div class="footer-inner">
        <div class="footer-brand">
          <h3>✦ 星络客 StarNet CRM</h3>
          <p>星络相连，六度通达</p>
        </div>
        <div class="footer-links">
          <dl>
            <dt>产品</dt>
            <dd><a>联系人管理</a></dd>
            <dd><a>关系图谱</a></dd>
            <dd><a>路径搜索</a></dd>
            <dd><a>引荐管理</a></dd>
          </dl>
          <dl>
            <dt>方案</dt>
            <dd><a>销售团队</a></dd>
            <dd><a>商务拓展</a></dd>
            <dd><a>猎头招聘</a></dd>
            <dd><a>投资关系</a></dd>
          </dl>
          <dl>
            <dt>支持</dt>
            <dd><a>帮助中心</a></dd>
            <dd><a>API文档</a></dd>
            <dd><a>联系我们</a></dd>
          </dl>
        </div>
      </div>
      <div class="footer-bottom">
        <span>&copy; 2026 星络客 StarNet CRM. All rights reserved.</span>
      </div>
    </footer>

    <!-- ─── 登录弹窗 ─── -->
    <a-modal
      v-model:open="showLogin"
      title="欢迎回来"
      :footer="null"
      :width="420"
      centered
      destroyOnClose
    >
      <a-form :model="loginForm" layout="vertical" @finish="handleLogin">
        <a-form-item label="用户名" name="username" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model:value="loginForm.username" size="large" placeholder="请输入用户名" />
        </a-form-item>
        <a-form-item label="密码" name="password" :rules="[{ required: true, message: '请输入密码' }]">
          <a-input-password v-model:value="loginForm.password" size="large" placeholder="请输入密码" />
        </a-form-item>
        <a-button type="primary" html-type="submit" size="large" block :loading="loginLoading">
          登录
        </a-button>
      </a-form>
      <div class="modal-switch">
        还没有账号？<a @click="showLogin = false; showRegister = true">立即注册</a>
      </div>
    </a-modal>

    <!-- ─── 注册弹窗 ─── -->
    <a-modal
      v-model:open="showRegister"
      title="创建账号"
      :footer="null"
      :width="460"
      centered
      destroyOnClose
    >
      <a-form :model="registerForm" layout="vertical" @finish="handleRegister">
        <a-form-item label="姓名" name="name" :rules="[{ required: true, message: '请输入姓名' }]">
          <a-input v-model:value="registerForm.name" size="large" placeholder="请输入真实姓名" />
        </a-form-item>
        <a-form-item label="公司名称" name="company">
          <a-input v-model:value="registerForm.company" size="large" placeholder="请输入公司名称（选填）" />
        </a-form-item>
        <a-form-item label="用户名" name="username" :rules="[{ required: true, message: '请输入用户名' }]">
          <a-input v-model:value="registerForm.username" size="large" placeholder="用于登录的用户名" />
        </a-form-item>
        <a-form-item label="密码" name="password" :rules="[{ required: true, min: 6, message: '密码至少6位' }]">
          <a-input-password v-model:value="registerForm.password" size="large" placeholder="设置登录密码" />
        </a-form-item>
        <a-form-item label="手机号" name="phone">
          <a-input v-model:value="registerForm.phone" size="large" placeholder="请输入手机号（选填）" />
        </a-form-item>
        <a-button type="primary" html-type="submit" size="large" block :loading="regLoading">
          免费注册
        </a-button>
      </a-form>
      <div class="modal-switch">
        已有账号？<a @click="showRegister = false; showLogin = true">去登录</a>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { CheckOutlined, TeamOutlined, ApartmentOutlined, SearchOutlined, SendOutlined, BarChartOutlined, TagsOutlined } from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

// ─── 导航滚动效果 ───
const isScrolled = ref(false)
const onScroll = () => { isScrolled.value = window.scrollY > 60 }
onMounted(() => window.addEventListener('scroll', onScroll))
onUnmounted(() => window.removeEventListener('scroll', onScroll))

// ─── 登录 ───
const showLogin = ref(false)
const loginLoading = ref(false)
const loginForm = reactive({ username: '', password: '' })

async function handleLogin() {
  loginLoading.value = true
  try {
    await auth.login(loginForm.username, loginForm.password)
    message.success('登录成功')
    showLogin.value = false
    router.push('/dashboard')
  } catch {
    // error handled by interceptor
  } finally {
    loginLoading.value = false
  }
}

// ─── 注册 ───
const showRegister = ref(false)
const regLoading = ref(false)
const registerForm = reactive({
  name: '',
  company: '',
  username: '',
  password: '',
  phone: '',
})

async function handleRegister() {
  regLoading.value = true
  try {
    await auth.register({
      username: registerForm.username,
      password: registerForm.password,
      name: registerForm.name,
      phone: registerForm.phone,
    })
    message.success('注册成功')
    showRegister.value = false
    router.push('/dashboard')
  } catch {
    // error handled by interceptor
  } finally {
    regLoading.value = false
  }
}

// ─── 锚点滚动 ───
function scrollTo(id: string) {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}
function scrollToTop() {
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

// ─── 功能卡片 ───
const features = [
  { icon: TeamOutlined, title: '联系人管理', desc: '多维度记录和管理人脉信息，支持标签分类、智能搜索和批量导入导出', bg: 'linear-gradient(135deg, #2B5FD7, #4A90D9)' },
  { icon: ApartmentOutlined, title: '关系图谱', desc: '力导向图直观展示人脉网络，一键查看1度关系和关系强度', bg: 'linear-gradient(135deg, #00C9A7, #36CFC9)' },
  { icon: SearchOutlined, title: '六度路径搜索', desc: '基于图算法计算两个节点之间的最短社交路径，发现隐藏的人脉通道', bg: 'linear-gradient(135deg, #F5A623, #FFC53D)' },
  { icon: SendOutlined, title: '引荐管理', desc: '结构化管理引荐流程，从发起请求到最终建立联系，全程追踪', bg: 'linear-gradient(135deg, #FF6B6B, #FF8787)' },
  { icon: BarChartOutlined, title: '人脉分析', desc: '多维度数据分析报表，包括关系类型分布、人脉增长趋势、超级连接者识别', bg: 'linear-gradient(135deg, #722ED1, #9254DE)' },
  { icon: TagsOutlined, title: '标签体系', desc: '自定义标签库，灵活分类人脉，便于精准检索和批量运营', bg: 'linear-gradient(135deg, #13C2C2, #36CFC9)' },
]

// ─── 定价方案 ───
const plans = [
  {
    name: '免费版',
    price: 0,
    desc: '适合个人和小团队体验',
    features: ['100个人脉上限', '基础关系图谱', '标签管理', '单租户', '社区支持'],
    cta: '免费开始',
    featured: false,
  },
  {
    name: '专业版',
    price: 29,
    desc: '适合成长中的销售和BD团队',
    features: ['无限人脉关系', '完整关系图谱', '六度路径搜索', '引荐工作流', '数据分析报表', 'API 接口', '邮件支持'],
    cta: '免费试用14天',
    featured: true,
  },
  {
    name: '企业版',
    price: 59,
    desc: '适合大型组织和重度用户',
    features: ['专业版全部功能', '多租户管理', '自定义数据字段', 'SSO单点登录', '私有化部署', '专属客户成功经理', '7×24小时支持'],
    cta: '联系我们',
    featured: false,
  },
]
</script>

<style scoped lang="less">
// ─── 变量 ───
@primary: #2B5FD7;
@primary-light: #4A7AE8;
@success: #00C9A7;
@gold: #F5A623;
@bg: #F5F7FA;
@text: #1A1A1A;
@text-secondary: #666;
@white: #fff;

// ─── 导航 ───
.nav-header {
  position: fixed; top: 0; left: 0; right: 0; z-index: 1000;
  padding: 0 40px; height: 64px;
  display: flex; align-items: center;
  background: rgba(255,255,255,0.92); backdrop-filter: blur(12px);
  transition: box-shadow 0.3s;
  &.scrolled { box-shadow: 0 2px 16px rgba(0,0,0,0.08); }
}
.nav-inner {
  width: 100%; max-width: 1200px; margin: 0 auto;
  display: flex; align-items: center; justify-content: space-between;
}
.nav-brand {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  .brand-icon { font-size: 28px; color: @primary; }
  .brand-text { font-size: 20px; font-weight: 700; color: @text; }
  .brand-sub { font-size: 12px; color: @text-secondary; margin-top: 4px; }
}
.nav-links {
  display: flex; align-items: center; gap: 24px;
  a { color: @text-secondary; cursor: pointer; font-size: 14px; transition: color 0.2s;
    &:hover { color: @primary; }
  }
}

// ─── HERO ───
.hero {
  position: relative; min-height: 680px;
  display: flex; align-items: center; justify-content: center;
  overflow: hidden; padding-top: 64px;
}
.hero-bg {
  position: absolute; inset: 0;
  background: linear-gradient(135deg, #f0f5ff 0%, #e8f0ff 30%, #f5f7fa 60%, #e6f9f5 100%);
  &::before {
    content: ''; position: absolute; top: -50%; right: -10%;
    width: 600px; height: 600px; border-radius: 50%;
    background: radial-gradient(circle, rgba(43,95,215,0.08), transparent 70%);
  }
  &::after {
    content: ''; position: absolute; bottom: -30%; left: -5%;
    width: 400px; height: 400px; border-radius: 50%;
    background: radial-gradient(circle, rgba(0,201,167,0.06), transparent 70%);
  }
}
.hero-content {
  position: relative; z-index: 1; text-align: center; max-width: 800px; padding: 0 24px;
}
.hero-badge {
  display: inline-block; padding: 6px 20px; border-radius: 20px;
  background: rgba(43,95,215,0.1); color: @primary;
  font-size: 14px; margin-bottom: 24px;
}
.hero-title {
  font-size: 52px; font-weight: 800; line-height: 1.3; color: @text; margin: 0 0 20px;
  .highlight { background: linear-gradient(135deg, @primary, @success); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
}
.hero-desc {
  font-size: 17px; color: @text-secondary; line-height: 1.8; margin: 0 0 36px;
}
.hero-actions {
  display: flex; gap: 16px; justify-content: center; margin-bottom: 48px;
  .btn-cta { height: 48px; padding: 0 36px; font-size: 16px; border-radius: 8px; }
  .btn-demo { height: 48px; padding: 0 36px; font-size: 16px; border-radius: 8px; border-color: #d9d9d9; }
}
.hero-stats {
  display: flex; align-items: center; justify-content: center; gap: 40px;
}
.stat-item {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  .stat-num { font-size: 32px; font-weight: 700; color: @text; }
  .stat-label { font-size: 13px; color: @text-secondary; }
}
.stat-divider { width: 1px; height: 40px; background: #ddd; }

// ─── 通用区块头 ───
.section-head {
  text-align: center; margin-bottom: 48px;
  h2 { font-size: 36px; font-weight: 700; color: @text; margin: 0 0 12px; }
  p { font-size: 16px; color: @text-secondary; margin: 0; }
  &.light {
    h2 { color: @white; }
    p { color: rgba(255,255,255,0.7); }
  }
}

// ─── 功能 ───
.features {
  padding: 80px 40px; max-width: 1200px; margin: 0 auto;
}
.feature-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px;
}
.feature-card {
  padding: 32px 28px; border-radius: 12px; background: @white;
  border: 1px solid #eee; transition: all 0.3s; cursor: default;
  &:hover { transform: translateY(-4px); box-shadow: 0 12px 32px rgba(0,0,0,0.08); border-color: @primary-light; }
  .feature-icon {
    width: 56px; height: 56px; border-radius: 14px; display: flex; align-items: center; justify-content: center;
    margin-bottom: 20px; font-size: 24px; color: @white;
  }
  h3 { font-size: 18px; font-weight: 600; margin: 0 0 8px; color: @text; }
  p { font-size: 14px; color: @text-secondary; line-height: 1.7; margin: 0; }
}

// ─── 展示 ───
.showcase {
  padding: 80px 40px; background: linear-gradient(135deg, #1a1a2e, #16213e);
}
.showcase-grid {
  max-width: 1080px; margin: 0 auto;
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 28px;
}
.showcase-item {
  text-align: center;
  .showcase-img {
    height: 180px; border-radius: 12px; margin-bottom: 20px;
    background: rgba(255,255,255,0.06); border: 1px solid rgba(255,255,255,0.1);
    position: relative; overflow: hidden;
    &::after {
      content: ''; position: absolute; inset: 0;
      background: linear-gradient(135deg, rgba(43,95,215,0.15), rgba(0,201,167,0.08));
    }
    &.img-graph::before { content: '●'; position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); font-size: 48px; color: rgba(43,95,215,0.4); }
    &.img-path::before { content: '⤳'; position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); font-size: 48px; color: rgba(0,201,167,0.4); }
    &.img-stats::before { content: '▣'; position: absolute; top: 50%; left: 50%; transform: translate(-50%,-50%); font-size: 48px; color: rgba(245,166,35,0.4); }
  }
  h4 { font-size: 18px; font-weight: 600; color: @white; margin: 0 0 8px; }
  p { font-size: 14px; color: rgba(255,255,255,0.6); margin: 0; line-height: 1.6; }
}

// ─── 定价 ───
.pricing {
  padding: 80px 40px; max-width: 1200px; margin: 0 auto;
}
.pricing-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px;
}
.pricing-card {
  padding: 36px 28px; border-radius: 16px; background: @white;
  border: 2px solid #eee; text-align: center; position: relative;
  transition: all 0.3s;
  &:hover { transform: translateY(-4px); box-shadow: 0 12px 32px rgba(0,0,0,0.06); }
  &.featured { border-color: @primary; box-shadow: 0 8px 32px rgba(43,95,215,0.12); }
  h3 { font-size: 20px; font-weight: 700; margin: 0 0 16px; }
  .price { margin-bottom: 8px;
    .currency { font-size: 20px; font-weight: 600; vertical-align: top; color: @text; }
    .amount { font-size: 48px; font-weight: 800; color: @primary; }
    .period { font-size: 14px; color: @text-secondary; }
  }
  .plan-desc { font-size: 13px; color: @text-secondary; margin: 0 0 24px; }
  .plan-features {
    list-style: none; padding: 0; margin: 0 0 28px; text-align: left;
    li { padding: 8px 0; font-size: 14px; color: @text-secondary; display: flex; align-items: center; gap: 8px;
      .anticon { color: @success; }
    }
  }
}
.popular-badge {
  position: absolute; top: -12px; left: 50%; transform: translateX(-50%);
  background: linear-gradient(135deg, @primary, @primary-light);
  color: @white; padding: 4px 20px; border-radius: 12px; font-size: 12px; font-weight: 600;
}

// ─── 案例 ───
.cases {
  padding: 80px 40px; background: @bg;
}
.case-logos {
  max-width: 800px; margin: 0 auto;
  display: flex; flex-wrap: wrap; gap: 12px; justify-content: center;
  span {
    padding: 10px 24px; border-radius: 8px; background: @white;
    color: @text-secondary; font-size: 14px; font-weight: 500;
    border: 1px solid #eee;
  }
}

// ─── 页脚 ───
.footer {
  background: #1a1a2e; padding: 48px 40px 24px; color: rgba(255,255,255,0.7);
  .footer-inner { max-width: 1200px; margin: 0 auto; display: flex; justify-content: space-between; gap: 60px; margin-bottom: 32px; }
  .footer-brand {
    h3 { font-size: 18px; color: @white; margin: 0 0 8px; }
    p { margin: 0; font-size: 14px; }
  }
  .footer-links { display: flex; gap: 60px;
    dl { margin: 0;
      dt { font-size: 14px; font-weight: 600; color: @white; margin-bottom: 12px; }
      dd { margin: 0 0 8px; font-size: 13px;
        a { color: rgba(255,255,255,0.5); cursor: pointer; transition: color 0.2s;
          &:hover { color: @white; }
        }
      }
    }
  }
  .footer-bottom { max-width: 1200px; margin: 0 auto; padding-top: 20px; border-top: 1px solid rgba(255,255,255,0.1); text-align: center; font-size: 13px; }
}

// ─── 弹窗切换链接 ───
.modal-switch { text-align: center; margin-top: 16px; font-size: 13px; color: @text-secondary;
  a { color: @primary; cursor: pointer; }
}

// ─── 响应式 ───
@media (max-width: 768px) {
  .nav-links a:not(.ant-btn) { display: none; }
  .hero-title { font-size: 32px; }
  .feature-grid, .pricing-grid, .showcase-grid { grid-template-columns: 1fr; }
  .hero-stats { flex-direction: column; gap: 16px; }
  .stat-divider { display: none; }
  .footer-inner { flex-direction: column; gap: 32px; }
  .footer-links { flex-wrap: wrap; gap: 32px; }
}
</style>

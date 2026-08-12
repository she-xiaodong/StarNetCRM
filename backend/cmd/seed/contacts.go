package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/starnet/crm/internal/model"
	repomysql "github.com/starnet/crm/internal/repository/mysql"
)

func seedContactsForUser(username string) error {
	// 查找用户
	var user model.User
	if err := repomysql.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return fmt.Errorf("用户 %s 不存在: %w", username, err)
	}

	// 检查是否已有种子数据
	var count int64
	repomysql.DB.Model(&model.Contact{}).Where("created_by = ?", user.ID).Count(&count)
	if count >= 356 {
		fmt.Printf("用户 %s 已有 %d 条联系人数据，跳过种子录入\n", username, count)
		return nil
	}

	fmt.Printf("为用户 %s (ID: %s) 生成356条联系人数据...\n", username, user.ID)

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// ─── 数据生成器 ───
	lastNames := []string{"张", "王", "李", "赵", "刘", "陈", "杨", "黄", "周", "吴", "徐", "孙", "马", "朱", "胡", "郭", "何", "高", "林", "罗", "郑", "梁", "谢", "宋", "唐", "许", "韩", "冯", "邓", "曹", "彭", "曾", "萧", "田", "董", "潘", "袁", "蔡", "蒋", "余", "于", "杜", "叶", "程", "苏", "魏", "吕", "丁", "任", "沈", "姚", "卢", "姜", "崔", "钟", "谭", "陆", "汪", "范", "金", "石", "廖", "贾", "夏", "韦", "付", "方", "白", "邹", "孟", "熊", "秦", "邱", "江", "尹", "薛", "闫", "段", "雷", "侯", "龙", "史", "陶", "黎", "贺", "顾", "毛", "郝", "龚", "邵", "万", "钱", "严", "覃", "武", "戴", "莫", "孔", "向", "汤"}
	firstNameChars := []string{"伟", "芳", "娜", "敏", "静", "丽", "强", "磊", "军", "洋", "勇", "艳", "杰", "娟", "涛", "明", "超", "秀英", "霞", "平", "刚", "桂英", "文", "华", "飞", "玉兰", "斌", "玲", "云", "国", "波", "志强", "秀", "玉", "建华", "睿", "宇", "欣", "雪", "思远", "雨桐", "智博", "一鸣", "天", "子涵", "梓轩", "皓", "昊然", "晓", "卓", "逸", "晨曦", "瑶", "博文", "星宇", "铭", "芷若", "嘉", "菁", "怡", "诗涵", "景", "恒", "宸", "彦", "彦博", "翔", "瑞"}

	companies := []string{
		"腾讯科技", "阿里巴巴", "字节跳动", "华为技术", "百度在线", "小米集团", "京东集团", "美团点评",
		"网易网络", "滴滴出行", "新浪微博", "快手科技", "商汤科技", "旷视科技", "科大讯飞", "海康威视",
		"大疆创新", "蔚来汽车", "理想汽车", "小鹏汽车", "招商银行", "平安集团", "中信证券", "华泰证券",
		"易方达基金", "红杉资本", "高瓴资本", "启明创投", "IDG资本", "经纬中国",
		"联影医疗", "迈瑞医疗", "药明康德", "华大基因", "中芯国际", "长江存储", "京东方", "三一重工",
		"美的集团", "格力电器", "海尔集团", "康佳集团", "中金公司", "中信建投", "国泰君安", "招商证券",
		"万科集团", "碧桂园", "龙湖集团", "华润置地", "港交所", "上交所", "深交所", "云锋基金",
		"顺丰速运", "中通快递", "圆通速递", "韵达快递", "极兔速递", "菜鸟网络", "德邦物流",
		"携程旅行", "去哪儿网", "同程旅行", "马蜂窝", "途家民宿", "飞猪旅行",
		"好未来教育", "新东方", "猿辅导", "作业帮", "网易有道", "高途教育", "中公教育",
		"光线传媒", "华谊兄弟", "博纳影业", "阿里影业", "猫眼娱乐", "阅文集团",
		"蚂蚁集团", "微众银行", "网商银行", "陆金所", "度小满金融", "众安保险", "水滴公司",
		"商汤科技", "旷视科技", "依图科技", "云从科技", "第四范式", "地平线", "寒武纪",
		"微软中国", "谷歌中国", "亚马逊中国", "苹果中国", "英特尔中国", "英伟达中国", "特斯拉中国",
	}

	titles := []string{
		"CEO", "CTO", "COO", "CFO", "CMO", "VP/副总裁", "总监", "高级经理", "经理", "主管",
		"架构师", "技术专家", "产品负责人", "项目负责人", "区域经理", "销售总监", "客户总监",
		"投资总监", "研究总监", "运营总监", "市场总监", "HR总监", "财务总监",
		"资深工程师", "高级设计师", "资深产品经理", "商务拓展总监", "战略总监", "合伙人",
	}

	departments := []string{"技术部", "产品部", "销售部", "市场部", "运营部", "财务部", "人力资源部", "研发部", "商务拓展部", "法务部", "行政部", "战略投资部", "数据部", "设计部"}

	tagNames := []string{"重要客户", "合作伙伴", "前同事", "同学", "校友", "投资人", "供应商", "行业大咖", "技术大牛", "销售渠道", "媒体关系", "政府关系", "导师", "创业伙伴", "猎头联系"}

	phonePrefixes := []string{"138", "139", "136", "137", "158", "159", "185", "186", "188", "189", "135", "150", "151", "152", "156", "157", "183", "187"}
	emailDomains := []string{
		"qq.com", "163.com", "126.com", "gmail.com", "outlook.com", "foxmail.com", "sina.com", "aliyun.com",
		"tencent.com", "alibaba-inc.com", "bytedance.com", "huawei.com", "baidu.com", "meituan.com", "jd.com",
	}

	// ─── 生成标签 ───
	tagMap := make(map[string]string)
	for _, name := range tagNames {
		tag := model.Tag{
			ID:       uuid.New().String(),
			Name:     name,
			TenantID: user.TenantID,
			Color:    randomTagColor(rng),
		}
		repomysql.DB.Where("name = ? AND tenant_id = ?", name, user.TenantID).FirstOrCreate(&tag)
		tagMap[name] = tag.ID
	}

	// ─── 批量生成联系人 ───
	batchSize := 50
	contacts := make([]model.Contact, 0, batchSize)
	total := 356

	for i := 0; i < total; i++ {
		name := randomChineseName(rng, lastNames, firstNameChars)
		company := companies[rng.Intn(len(companies))]
		title := titles[rng.Intn(len(titles))]
		dept := departments[rng.Intn(len(departments))]
		phone := randomPhone(rng, phonePrefixes)
		email := randomEmail(rng, name, emailDomains)

		// 随机选 2-4 个标签
		nTags := 2 + rng.Intn(3)
		perm := rng.Perm(len(tagNames))
		tagIDs := make([]string, 0, nTags)
		for j := 0; j < nTags && j < len(perm); j++ {
			tagIDs = append(tagIDs, tagMap[tagNames[perm[j]]])
		}
		tagsJSON, _ := json.Marshal(tagIDs)

		personID := uuid.New().String()
		contact := model.Contact{
			ID:         uuid.New().String(),
			TenantID:   user.TenantID,
			PersonID:   personID,
			Name:       name,
			Company:    company,
			Title:      title,
			Department: dept,
			Phone:      phone,
			Email:      email,
			Tags:       string(tagsJSON),
			CreatedBy:  user.ID,
		}

		contacts = append(contacts, contact)

		// 批量写入
		if len(contacts) >= batchSize {
			if err := repomysql.DB.Create(&contacts).Error; err != nil {
				return fmt.Errorf("写入联系人失败: %w", err)
			}
			fmt.Printf("  已录入 %d/%d 条...\n", i+1, total)
			contacts = contacts[:0]
		}
	}

	// 写入剩余
	if len(contacts) > 0 {
		if err := repomysql.DB.Create(&contacts).Error; err != nil {
			return fmt.Errorf("写入联系人失败: %w", err)
		}
		fmt.Printf("  已录入 %d/%d 条...\n", total, total)
	}

	fmt.Printf("✅ 完成！已为用户 %s 录入 %d 条联系人关系数据\n", username, total)
	return nil
}

// ─── Helper functions ───

func randomChineseName(rng *rand.Rand, lastNames, firstNameChars []string) string {
	ln := lastNames[rng.Intn(len(lastNames))]
	fn := firstNameChars[rng.Intn(len(firstNameChars))]
	// 30% 概率单字名
	if rng.Float32() < 0.3 {
		return ln + firstNameChars[rng.Intn(len(firstNameChars))]
	}
	return ln + fn
}

func randomPhone(rng *rand.Rand, prefixes []string) string {
	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := fmt.Sprintf("%08d", rng.Intn(100000000))
	return prefix + suffix
}

func randomEmail(rng *rand.Rand, name string, domains []string) string {
	// 拼音化名字
	pinyin := ""
	for _, r := range name {
		pinyin += string(r)
	}
	domain := domains[rng.Intn(len(domains))]
	return fmt.Sprintf("%s%d@%s", pinyin[3:], rng.Intn(9999), domain)
}

func randomNote(rng *rand.Rand, name, company string) string {
	notes := []string{
		fmt.Sprintf("在行业峰会上认识，%s的%s，对AI方向很感兴趣", company, name),
		fmt.Sprintf("通过%s王总介绍认识，合作过两次项目", company),
		fmt.Sprintf("%s的技术负责人，技术能力很强，值得深入交流", company),
		fmt.Sprintf("校友，微信经常互动，关系很好"),
		fmt.Sprintf("上个月在深圳见过面，约了下个季度喝咖啡"),
		fmt.Sprintf("投资圈的资源人，认识很多LP"),
		fmt.Sprintf("前同事推荐，据说%s近期有融资需求", company),
		fmt.Sprintf("行业老炮，对市场趋势判断很准"),
		fmt.Sprintf("性格开朗，喜欢打高尔夫，可以作为运动类社交对象"),
		fmt.Sprintf("%s的核心决策者之一，需要重点维护关系", company),
		fmt.Sprintf("在领英上认识的，对数据安全领域有深刻见解"),
		fmt.Sprintf("供应商联系人，合作了3年，非常可靠"),
		fmt.Sprintf("之前项目的合作伙伴，现在去了%s", company),
		fmt.Sprintf("参加过同一期EMBA课程，每季度聚会一次"),
	}
	return notes[rng.Intn(len(notes))]
}

func randomTagColor(rng *rand.Rand) string {
	colors := []string{"#2B5FD7", "#00C9A7", "#F5A623", "#FF6B6B", "#722ED1", "#13C2C2", "#FA8C16", "#EB2F96", "#1890FF", "#52C41A"}
	return colors[rng.Intn(len(colors))]
}

func randomDate(rng *rand.Rand, maxDaysAgo int) time.Time {
	days := rng.Intn(maxDaysAgo)
	return time.Now().Add(-time.Duration(days) * 24 * time.Hour)
}

package config

import (
	"flag"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Server              string
	Port                int
	User                string
	Password            string
	Interval            float64
	IsVnstat            bool
	CU                  string
	CT                  string
	CM                  string
	ProbePort           int
	ProbeProtocolPrefer string
	Debug               bool
	EnableTLS           bool
	TLSSkipVerify       bool
	TLSServerName       string
	Iface               string
	ExcludeNet          string
}

func LoadConfig() *Config {
	host := flag.String("host", "", "主机地址")
	port := flag.Int("port", 35601, "主机端口")
	user := flag.String("user", "", "客户端用户名")
	password := flag.String("password", "", "客户端密码")
	interval := flag.Float64("interval", 1.0, "数据发送间隔(秒)")
	dsn := flag.String("dsn", "", "DSN 格式: username:password@host:port")
	vnstat := flag.Bool("vnstat", false, "使用 vnstat 获取网络流量(仅Linux)")
	iface := flag.String("iface", "", "指定监控的网络接口名(多个用逗号分隔, 支持 auto 自动识别默认出口网卡)")
	excludeNet := flag.String("exclude-net", "", "排除的网络接口正则模式(如 'lan.*|wlan.*')")
	cu := flag.String("cu", "cu.tz.cloudcpp.com", "CU 探针地址")
	ct := flag.String("ct", "ct.tz.cloudcpp.com", "CT 探针地址")
	cm := flag.String("cm", "cm.tz.cloudcpp.com", "CM 探针地址")
	probePort := flag.Int("probePort", 80, "探针端口")
	proto := flag.String("proto", "ipv4", "探针协议偏好(ipv4或ipv6)")
	debug := flag.Bool("debug", false, "开启调试模式")
	tls := flag.Bool("tls", false, "启用 TLS 加密连接")
	tlsSkipVerify := flag.Bool("tls-skip-verify", false, "跳过 TLS 证书验证")
	tlsSni := flag.String("tls-sni", "", "自定义 TLS SNI 域名(默认为 host)")

	flag.Parse()

	cfg := &Config{
		Server:              *host,
		Port:                *port,
		User:                *user,
		Password:            *password,
		Interval:            *interval,
		IsVnstat:            *vnstat,
		Iface:               strings.TrimSpace(*iface),
		ExcludeNet:          strings.TrimSpace(*excludeNet),
		CU:                  *cu,
		CT:                  *ct,
		CM:                  *cm,
		ProbePort:           *probePort,
		ProbeProtocolPrefer: *proto,
		Debug:               *debug,
		EnableTLS:           *tls,
		TLSSkipVerify:       *tlsSkipVerify,
		TLSServerName:       *tlsSni,
	}

	if *dsn != "" {
		cfg.parseDSN(*dsn)
	}

	cfg.validate()
	return cfg
}

func (c *Config) parseDSN(dsn string) {
	parts := strings.Split(dsn, "@")
	if len(parts) != 2 {
		log.Fatal("DSN 格式错误, 缺少 @ 符号, 应为 username:password@host:port")
	}
	auth := strings.Split(parts[0], ":")
	if len(auth) != 2 {
		log.Fatal("DSN 格式错误, 缺少 : 号符, 应为 username:password@host:port")
	}
	c.User = auth[0]
	c.Password = auth[1]

	addr := strings.Split(parts[1], ":")
	c.Server = addr[0]
	if len(addr) == 2 {
		port, err := strconv.Atoi(addr[1])
		if err == nil {
			c.Port = port
		}
	}
}

func (c *Config) validate() {
	if c.Port < 1 || c.Port > 65535 {
		log.Fatal("端口号必须在1到65535之间")
	}
	if c.Server == "" || c.User == "" || c.Password == "" {
		log.Fatal("主机地址、用户名和密码不能为空")
	}

	if c.EnableTLS && c.TLSServerName == "" {
		c.TLSServerName = c.Server
	}
	
	switch strings.ToLower(c.ProbeProtocolPrefer) {
	case "ipv4":
		c.ProbeProtocolPrefer = "ip4"
	case "ipv6":
		c.ProbeProtocolPrefer = "ip6"
	default:
		c.ProbeProtocolPrefer = "ip"
	}

	if c.ExcludeNet != "" {
		if _, err := regexp.Compile(c.ExcludeNet); err != nil {
			log.Fatalf("exclude-net 正则表达式无效: %v", err)
		}
	}
}

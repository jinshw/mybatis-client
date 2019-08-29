package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/config"
	"gopkg.in/urfave/cli.v1"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

type Config struct {
	db              string
	table           string
	packageJavaBean string
	packageService  string
	packageDao      string
	packageMapper   string
}

var (
	DATA_SOURCE_NAME = "root:root@tcp(127.0.0.1:3306)/mountain?charset=utf8"
	CONFIG_FILE      = "./config/mysql-config.ini"
	TEMPLATE_PATH    = "./mysql/template"
	OUT_PATH         = "./mysql/out/"
	DB               = "mountain"
	TABLE            = "sys_role"
	PACKAGE_JAVABEAN = "com.site.mountain.entity"
	PACKAGE_SERVICE  = ""
	PACKAGE_DAO      = "com.site.mountain.dao.test2"
	MAPPER_PATH      = "com.site.mountain.dao.mapper"
)

var relationType = make(map[string]string)

func initRelationType() {
	relationType["int"] = "java.lang.Integer"
	relationType["varchar"] = "java.lang.String"
	relationType["char"] = "java.lang.String"
	relationType["blob"] = "java.lang.byte[]"
	relationType["text"] = "java.lang.String"
	relationType["integer"] = "java.lang.Long"
	relationType["tinyint"] = "java.lang.Integer"
	relationType["smallint"] = "java.lang.Integer"
	relationType["mediumint"] = "java.lang.Integer"
	relationType["bit"] = "java.lang.Boolean"
	relationType["bigint"] = "java.math.BigInteger"
	relationType["float"] = "java.lang.Float"
	relationType["double"] = "java.lang.Double"
	relationType["date"] = "java.sql.Date"
	relationType["time"] = "java.sql.Time"
	relationType["datetime"] = "java.sql.Timestamp"
	relationType["timestamp"] = "java.sql.Timestamp"
	relationType["year"] = "java.sql.Date"
	//扩展
	relationType["list"] = "java.util.List"
}

func initConfig() {
	c, err := config.ReadDefault(CONFIG_FILE)
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	TEMPLATE_PATH, _ = c.String("template", "TEMPLATE_PATH")
	OUT_PATH, _ = c.String("template", "OUT_PATH")
	DATA_SOURCE_NAME, _ = c.String("mysql", "DATA_SOURCE_NAME")
	DB, _ = c.String("mysql", "DB")
	TABLE, _ = c.String("mysql", "TABLE")
	PACKAGE_JAVABEAN, _ = c.String("package", "PACKAGE_JAVABEAN")
	PACKAGE_DAO, _ = c.String("package", "PACKAGE_DAO")
	MAPPER_PATH, _ = c.String("package", "MAPPER_PATH")

}

func main() {
	initConfig()
	initRelationType()

	db, err := sql.Open("mysql", DATA_SOURCE_NAME)
	if err != nil {
		log.Fatal(err)
	}
	//--start: 命令行实现
	app := cli.NewApp()
	app.Name = "Mybatis Genernator Tools"
	app.Usage = "mgt"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "table, t",
			Value: TABLE,
			Usage: "db table name,example:sys_user,sys_role",
		},
		cli.StringFlag{
			Name:  "packagejavabean, pj",
			Value: PACKAGE_JAVABEAN,
			Usage: "set java bean package",
		},
		cli.StringFlag{
			Name:  "packagedao, pd",
			Value: PACKAGE_DAO,
			Usage: "set dao package",
		},
		cli.StringFlag{
			Name:  "packagemapper, pm",
			Value: MAPPER_PATH,
			Usage: "set mapper package",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:     "packagejavabean",
			Aliases:  []string{"pj"},
			Usage:    "get java bean package",
			Category: "configuration",
			Action: func(c *cli.Context) error {
				fmt.Println("PACKAGE_JAVABEAN = ", PACKAGE_JAVABEAN)
				return nil
			},
		},
		{
			Name:     "packagedao",
			Aliases:  []string{"pd"},
			Usage:    "get dao package",
			Category: "configuration",
			Action: func(c *cli.Context) error {
				fmt.Println("PACKAGE_DAO = ", PACKAGE_DAO)
				return nil
			},
		},
		{
			Name:     "packagemapper",
			Aliases:  []string{"pm"},
			Usage:    "get mapper package",
			Category: "configuration",
			Action: func(c *cli.Context) error {
				fmt.Println("PACKAGE_MAPPER = ", MAPPER_PATH)
				return nil
			},
		},
	}
	app.Action = func(c *cli.Context) error {
		fmt.Println("******************")
		fmt.Println("Action...")
		PACKAGE_JAVABEAN = c.String("packagejavabean")
		PACKAGE_DAO = c.String("packagedao")
		MAPPER_PATH = c.String("packagemapper")
		TABLE = c.String("table")
		fmt.Println("PACKAGE_JAVABEAN=", PACKAGE_JAVABEAN)
		fmt.Println("PACKAGE_DAO=", PACKAGE_DAO)
		fmt.Println("MAPPER_PATH=", MAPPER_PATH)
		fmt.Println("TABLE=", TABLE)
		fmt.Println("******************")

		tables := strings.Split(TABLE, ",")
		for _, table := range tables {
			goMapperTools(db, table)
		}
		defer db.Close()

		return nil
	}
	app.Before = func(c *cli.Context) error {
		fmt.Println("app Before")
		return nil
	}
	app.After = func(c *cli.Context) error {
		fmt.Println("app After")
		return nil
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	cli.HelpFlag = cli.BoolFlag{
		Name:  "help, h",
		Usage: "Help!Help!",
	}

	cli.VersionFlag = cli.BoolFlag{
		Name:  "print-version, v",
		Usage: "print version",
	}
	errApp := app.Run(os.Args)
	if errApp != nil {
		log.Fatal(errApp)
	}
	//--end: 命令行实现


}

func goMapperTools(db *sql.DB, table string) {
	rows, _ := db.Query("select column_name,column_comment,data_type " +
		"from information_schema.columns " +
		"where table_name='" + table + "' and table_schema='" + DB + "'")

	columns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(columns))
	scans := make([]interface{}, len(columns))

	for i := range values {
		scans[i] = &values[i]
	}

	var result []map[string]string
	for rows.Next() {
		_ = rows.Scan(scans...)
		each := make(map[string]string)
		for i, col := range values {
			each[columns[i]] = string(col)
		}
		result = append(result, each)
	}
	//
	GetJavaBean(result, table)
	GetDaoFile(table)
	GetMapperFile(result, table)

	time.Sleep(10)
	defer rows.Close()
}

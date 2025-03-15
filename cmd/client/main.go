package main

import (
	"context"
	"fmt"
	"log"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/etcd"
	pb "siwuai/proto/article"
	pbcode "siwuai/proto/code"
	"time"

	"google.golang.org/grpc"
)

//// 用于模拟客户端通过gRPC调用LLM服务
//func main() {
//	// 连接到 gRPC 服务器（假设运行在 localhost:50051）
//	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
//	if err != nil {
//		log.Fatalf("无法连接到服务器: %v", err)
//	}
//	defer conn.Close()
//
//	// 设置请求上下文和超时
//	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
//	defer cancel()
//
//	client1 := pbcode.NewCodeServiceClient(conn)
//
//	req1 := &pbcode.CodeRequest{
//		CodeQuestion: "你是谁创造的",
//		UserId:       2,
//	}
//
//	resp1, err := client1.ExplainCode(ctx, req1)
//	if err != nil {
//		fmt.Println("client1.ExplainCode()", err)
//	}
//
//	fmt.Println("resp1:", resp1)
//
//}

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("configs")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// etcd 注册初始化，使用配置文件中的 etcd 配置
	etcdCfg := cfg.Etcd
	registry, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints, etcdCfg.ServiceName, etcdCfg.ServiceAddr, etcdCfg.TTL)
	if err != nil {
		log.Fatalf("创建 etcd 注册器失败: %v", err)
	}

	// 服务发现
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := registry.Discover(ctx, cfg.Etcd.ServiceName)
	if err != nil {
		log.Fatalf("服务发现失败: %v", err)
	}
	if len(addrs) == 0 {
		log.Fatalf("未找到可用的服务实例")
	}

	// 连接到 gRPC 服务器（假设运行在 localhost:50051）
	conn, err := grpc.Dial(addrs[0], grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到服务器: %v", err)
	}
	defer conn.Close()

	// 设置请求上下文和超时
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	client1 := pb.NewArticleServiceClient(conn)

	//tags := []string{"前端", "后端", "MySQL", "Redis", "Git", "HTTP", "Gorm"}
	//req1 := &pb.GetArticleInfoFirstRequest{
	//	Content: "一. First()\nresult := tx.Model(&models.Attachment{}).Where(\"home = ? AND home_id = ?\", attachment.Home, attachment.HomeID).First(&existingAttachment)\n1\n如果没有查询到数据，result.Error的值是什么？\n在使用 GORM（Go 的 ORM 框架）进行数据库操作时，result.Error 的值取决于查询的执行情况。\n如果查询没有找到任何数据，result.Error 的值通常是 gorm.ErrRecordNotFound。\n\n解释：\n1.First 方法：First 方法会尝试获取查询结果的第一条记录。如果没有找到任何匹配的记录，GORM 会返回一个错误，通常是 gorm.ErrRecordNotFound。\n\n2.result.Error：result.Error 是 GORM 的 Result 结构体中的一个字段，用于存储查询或操作过程中发生的错误。如果没有错误，它的值为 nil；如果有错误，它会包含具体的错误信息。\n\n示例代码：\nif result.Error != nil {\n    if errors.Is(result.Error, gorm.ErrRecordNotFound) {\n        // 没有找到数据\n        fmt.Println(\"No record found\")\n    } else {\n        // 其他错误\n        fmt.Println(\"Error occurred:\", result.Error)\n    }\n} else {\n    // 查询成功，数据存储在 existingAttachment 中\n    fmt.Println(\"Record found:\", existingAttachment)\n}\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12\n注意事项：\n如果你希望在没有找到数据时不返回错误，可以使用 Take 方法代替 First。Take 方法在没有找到数据时不会返回错误，而是将目标变量设置为默认值。\n如果你使用的是 GORM v2 或更高版本，gorm.ErrRecordNotFound 是一个常量错误，可以通过 errors.Is 来判断是否是“未找到记录”的错误。\n二. Create()\n// 向api表中添加api相关信息\nerr := tx.Model(models.Api{}).Create(&api).Error\nif err != nil {\ntx.Rollback() // 回滚事务\nreturn fmt.Errorf(\"CreateApiRep -> 向api表中添加api相关信息失败 -> %s\", err)\n}\n1\n2\n3\n4\n5\n6\n如果插入数据失败，err的值是什么?\n在使用 GORM 的 Create 方法插入数据时，如果插入操作失败，err 的值将是一个具体的错误对象，其内容取决于失败的原因。这些错误通常是由数据库操作或约束违反引起的。以下是一些常见的错误类型及其可能的值：\n\n1. 数据库约束违反\n如果插入的数据违反了数据库的约束（如主键冲突、唯一约束冲突、外键约束等），err 的值通常是一个 *pq.Error（PostgreSQL）或 *mysql.MySQLError（MySQL）对象，具体取决于你使用的数据库驱动。\n\n示例错误：\n- 主键冲突（MySQL）：\nError 1062: Duplicate entry '1' for key 'PRIMARY'\n\n- 唯一约束冲突（MySQL）：\nError 1062: Duplicate entry 'example' for key 'unique_key_name'\n\n- 外键约束冲突（MySQL）：\nError 1452: Cannot add or update a child row: a foreign key constraint fails\n1\n2\n3\n4\n5\n6\n7\n8\n示例错误（PostgreSQL）：\n- 主键冲突：\npq: duplicate key value violates unique constraint \"pkey_name\"\n\n- 唯一约束冲突：\npq: duplicate key value violates unique constraint \"unique_constraint_name\"\n1\n2\n3\n4\n5\n2. 数据类型不匹配\n如果插入的数据类型与数据库表定义不匹配（如字符串长度超出限制、字段类型不匹配等），err 的值通常是一个语法错误或类型错误。\n示例错误（MySQL）：\n\nError 1265: Data truncated for column 'column_name' at row 1\n1\n3. 数据库连接问题\n如果数据库连接失败或超时，err 的值可能是一个连接错误。\n示例错误：\n\ndial tcp 127.0.0.1:3306: connect: connection refused\n1\n4. 其他错误\n其他可能的错误包括：\n\nSQL 语法错误：\nError 1064: You have an error in your SQL syntax\n1\n事务回滚错误：\ntransaction is already rolled back\n1\n如何处理这些错误\n在实际开发中，你可以根据错误类型进行相应的处理。以下是一些常见的处理方式：\n\n示例代码：\nerr := tx.Model(models.Api{}).Create(&api).Error\nif err != nil {\n    tx.Rollback() // 回滚事务\n    if gorm.IsRecordNotFoundError(err) {\n        return fmt.Errorf(\"CreateApiRep -> 插入数据失败：记录未找到\")\n    } else if strings.Contains(err.Error(), \"Duplicate entry\") {\n        return fmt.Errorf(\"CreateApiRep -> 插入数据失败：唯一约束冲突\")\n    } else {\n        return fmt.Errorf(\"CreateApiRep -> 插入数据失败 -> %s\", err)\n    }\n}\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n总结\nerr 的值取决于插入操作失败的具体原因。\n常见的错误类型包括数据库约束违反、数据类型不匹配、连接问题等。\n你可以通过检查错误消息或使用 GORM 提供的工具（如 gorm.IsRecordNotFoundError）来处理这些错误。\n三. Update()\n// 更新 api_group 表中的 group_id\nresag := tx.Model(&models.ApiDictItemGroup{}).Where(\"api_id = ?\", api.ID).Update(\"group_id\", groupId)\nif resag.Error != nil {\ntx.Rollback() // 回滚事务\nreturn fmt.Errorf(\"UpdateApiRep -> 更新 api_group 表中的 group_id失败 -> %s\", resag.Error)\n}\n1\n2\n3\n4\n5\n6\n当更新的数据与原数据相同时，resag.RowsAffected的值是什么?\n在 GORM 中，Update 方法的 RowsAffected 属性表示实际被更新的行数。如果更新的数据与原数据相同，RowsAffected 的值通常为 0，表示没有行被实际修改。\n\n示例代码及解释\n假设你有一个 ApiDictItemGroup 模型，并试图更新其 group_id 字段：\n\nresag := tx.Model(&models.ApiDictItemGroup{}).Where(\"api_id = ?\", api.ID).Update(\"group_id\", groupId)\n1\n如果\ngroupId 与数据库中当前的 group_id 值相同：\nresag.Error 是 nil，表示没有错误。\nresag.RowsAffected 的值为 0，因为没有实际修改任何行。\n数据库不会执行实际的更新操作，因为没有数据变化。\n如果\ngroupId 与数据库中当前的 group_id 值不同：\nresag.Error 是 nil，表示没有错误。\nresag.RowsAffected 的值为 1，因为有一行数据被修改。\n结论\n当你尝试更新的数据与原数据相同时，RowsAffected 的值为 0。这表明没有行被实际修改，但操作是成功的。\n\n建议\n如果你需要确保某些逻辑仅在数据实际更改时执行，可以检查 resag.RowsAffected 的值是否大于 0。例如：\n\nif resag.RowsAffected > 0 {\n    // 数据被实际修改，执行某些逻辑\n}\n1\n2\n3\n四. Delete()\nresult2 := tx.Model(&models.ApiDictItemGroup{}).Where(\"api_id = ?\", id).Delete(nil)\nif result2.Error != nil {\ntx.Rollback() // 回滚事务\nreturn fmt.Errorf(\"DeleteApiRep ->  删除api_group表中的信息失败 -> %s\", result2.Error)\n}\n1\n2\n3\n4\n5\n如果要删除的数据不存在，result2.Error 的值是什么?\n在 GORM 中，如果要删除的数据不存在（即删除条件未匹配到任何记录），result2.Error 的值通常为 nil，而不是一个具体的错误。这是因为 GORM 认为“未找到匹配记录”并不是一个错误，而是一个正常的操作结果。\n\n在这种情况下，result2.RowsAffected 的值会是 0，表示没有行被删除。如果你需要根据是否删除了数据来判断操作是否成功，可以通过检查 result2.RowsAffected 来实现。\n\n示例代码\nresult2 := tx.Model(&models.ApiDictItemGroup{}).Where(\"api_id = ?\", id).Delete(nil)\nif result2.Error != nil {\n    tx.Rollback() // 回滚事务\n    return fmt.Errorf(\"DeleteApiRep -> 删除api_group表中的信息失败 -> %s\", result2.Error)\n}\n\nif result2.RowsAffected == 0 {\n    tx.Rollback() // 回滚事务\n    return fmt.Errorf(\"DeleteApiRep -> 删除api_group表中的信息失败 -> 未找到匹配的记录\")\n}\n1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n总结\n如果删除条件未匹配到任何记录，result2.Error 的值是 nil。\n如果需要判断是否删除了数据，可以通过检查 result2.RowsAffected 是否为 0 来实现。\n如果 result2.Error 不为 nil，则表示删除操作失败，可能是由于其他错误（如数据库连接问题、SQL 执行错误等）。\n获取以上文章的摘要和总结，并根据以下标签，给该篇文章匹配标签，标签如下：\n前端、后端、MySQL、redis、git、http、gorm。" +
	//		"回答中应仅仅只包含三部分: 摘要、总结、标签，其他多余部分都不要",
	//	Tags: tags,
	//}
	//resp1, err := client1.GetArticleInfoFirst(ctx, req1)
	//if err != nil {
	//	fmt.Printf("client1.GetArticleInfoFirst -------> \n %v \n %v", resp1, err)
	//} else {
	//	fmt.Printf("++++++++++++++++++> \n %v", resp1)
	//}

	//req1 := &pb.SaveArticleIDRequest{
	//	Key:       "2d90850917f6353024631f0b5d3b42cca6bef73525b5b9505e6e24e322799d06",
	//	ArticleID: 1,
	//}
	//
	//resp1, err := client1.SaveArticleID(ctx, req1)
	//if err != nil {
	//	fmt.Printf("client1.SaveArticleID -------> \n %v \n %v", resp1, err)
	//} else {
	//	fmt.Printf("++++++++++++++++++> \n %v", resp1)
	//}

	//req1 := &pb.GetArticleInfoRequest{
	//	ArticleID: 1,
	//}
	//resp1, err := client1.GetArticleInfo(ctx, req1)
	//if err != nil {
	//	fmt.Printf("client1.GetArticleInfo -------> \n %v \n %v", resp1, err)
	//} else {
	//	fmt.Printf("++++++++++++++++++> \n %v", resp1)
	//}

	req1 := &pb.DelArticleInfoRequest{
		ArticleID: 1,
	}
	resp1, err := client1.DelArticleInfo(ctx, req1)
	if err != nil {
		fmt.Printf("client1.DelArticleInfo -------> \n %v \n %v", resp1, err)
	} else {
		fmt.Printf("++++++++++++++++++> \n %v", resp1)
	}

	// code ------------------------------------------
	client2 := pbcode.NewCodeServiceClient(conn)

	req2 := &pbcode.CodeRequest{
		CodeQuestion: "你是谁创造的",
		UserId:       2,
		CodeType:     "go",
	}

	resp2, err := client2.ExplainCode(ctx, req2)
	if err != nil {
		fmt.Println("client1.ExplainCode()", err)
	}

	fmt.Println("resp2:", resp2)

}

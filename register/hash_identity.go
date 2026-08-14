package register

import (
	"fmt"

	"app/comm"
	"app/dao/model"

	"github.com/spf13/cobra"
	"github.com/zjutjh/mygo/ndb"
)

// hashIdentityCommand 将历史明文身份证转换为不可逆的 HMAC-SHA256 摘要。
// 该命令具有幂等性：64 位十六进制摘要不会被重复处理。
func hashIdentityCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "hash-identities",
		Short: "散列 peoples 表中已有的明文身份证号",
		RunE: func(cmd *cobra.Command, args []string) error {
			var people []model.People
			db := ndb.Pick()
			if err := db.Where("CHAR_LENGTH(TRIM(identity)) <> 64").Find(&people).Error; err != nil {
				return err
			}
			for i := range people {
				digest, err := comm.EncryptIdentity(people[i].Identity)
				if err != nil {
					return fmt.Errorf("散列人员 %d 的身份证失败: %w", people[i].ID, err)
				}
				if err := db.Model(&model.People{}).Where("id = ?", people[i].ID).Update("identity", digest).Error; err != nil {
					return fmt.Errorf("保存人员 %d 的身份证摘要失败: %w", people[i].ID, err)
				}
			}
			cmd.Printf("已散列 %d 条身份证记录\n", len(people))
			return nil
		},
	}
}

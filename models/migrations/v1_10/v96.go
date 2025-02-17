// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package v1_10 //nolint

import (
	"path/filepath"

	"code.gitea.io/gitea/modules/setting"
	"code.gitea.io/gitea/modules/util"

	"xorm.io/xorm"
)

func DeleteOrphanedAttachments(x *xorm.Engine) error {
	type Attachment struct {
		ID        int64  `xorm:"pk autoincr"`
		UUID      string `xorm:"uuid UNIQUE"`
		IssueID   int64  `xorm:"INDEX"`
		ReleaseID int64  `xorm:"INDEX"`
		CommentID int64
	}

	sess := x.NewSession()
	defer sess.Close()

	limit := setting.Database.IterateBufferSize
	if limit <= 0 {
		limit = 50
	}

	for {
		attachements := make([]Attachment, 0, limit)
		if err := sess.Where("`issue_id` = 0 and (`release_id` = 0 or `release_id` not in (select `id` from `release`))").
			Cols("id, uuid").Limit(limit).
			Asc("id").
			Find(&attachements); err != nil {
			return err
		}
		if len(attachements) == 0 {
			return nil
		}

		ids := make([]int64, 0, limit)
		for _, attachment := range attachements {
			ids = append(ids, attachment.ID)
		}
		if len(ids) > 0 {
			if _, err := sess.In("id", ids).Delete(new(Attachment)); err != nil {
				return err
			}
		}

		for _, attachment := range attachements {
			uuid := attachment.UUID
			if err := util.RemoveAll(filepath.Join(setting.Attachment.Path, uuid[0:1], uuid[1:2], uuid)); err != nil {
				return err
			}
		}
		if len(attachements) < limit {
			return nil
		}
	}
}

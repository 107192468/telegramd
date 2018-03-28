/*
 *  Copyright (c) 2017, https://github.com/nebulaim
 *  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rpc

import (
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/baselib/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"github.com/nebulaim/telegramd/biz/dal/dao"
)

// contacts.unblock#e54100bd id:InputUser = Bool;
func (s *ContactsServiceImpl) ContactsUnblock(ctx context.Context, request *mtproto.TLContactsUnblock) (*mtproto.Bool, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("ContactsUnblock - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	switch request.GetId().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputUserEmpty:
	case mtproto.TLConstructor_CRC32_inputUserSelf:
		dao.GetUserContactsDAO(dao.DB_MASTER).UpdateBlock(0, md.UserId, md.UserId)
	case mtproto.TLConstructor_CRC32_inputUser:
		// TODO(@benqi): Check InputUser's userId and access_hash
		dao.GetUserContactsDAO(dao.DB_MASTER).UpdateBlock(0, md.UserId, request.GetId().GetData2().GetUserId())
	}

	glog.Infof("ContactsUnblock - reply: {true}")
	return mtproto.ToBool(true), nil
}

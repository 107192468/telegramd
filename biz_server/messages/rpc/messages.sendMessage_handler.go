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
	"github.com/nebulaim/telegramd/base/logger"
	"github.com/nebulaim/telegramd/grpc_util"
	"github.com/nebulaim/telegramd/mtproto"
	"golang.org/x/net/context"
	"time"
	"github.com/nebulaim/telegramd/biz_model/base"
	"github.com/nebulaim/telegramd/biz_model/model"
	"github.com/nebulaim/telegramd/zproto"
	"fmt"
	"github.com/nebulaim/telegramd/biz_server/sync"
)

// 流程：
//  1. 入库: 1）存消息数据，2）分别存到发件箱和收件箱里
//  2. 离线推送
//  3. 在线推送
// messages.sendMessage#fa88427a flags:# no_webpage:flags.1?true silent:flags.5?true background:flags.6?true clear_draft:flags.7?true peer:InputPeer reply_to_msg_id:flags.0?int message:string random_id:long reply_markup:flags.2?ReplyMarkup entities:flags.3?Vector<MessageEntity> = Updates;
func (s *MessagesServiceImpl) MessagesSendMessage(ctx context.Context, request *mtproto.TLMessagesSendMessage) (*mtproto.Updates, error) {
	md := grpc_util.RpcMetadataFromIncoming(ctx)
	glog.Infof("MessagesSendMessage - metadata: %s, request: %s", logger.JsonDebugData(md), logger.JsonDebugData(request))

	var (
		sentMessage *mtproto.TLUpdateShortSentMessage
		err error
	)

	switch request.GetPeer().GetConstructor() {
	case mtproto.TLConstructor_CRC32_inputPeerSelf:
		sentMessage, err = sendPeerSelfMessage(md, request)
	case mtproto.TLConstructor_CRC32_inputPeerUser:
		sentMessage, err = sendPeerUserMessage(md, request)
	case mtproto.TLConstructor_CRC32_inputPeerChat:
		sentMessage, err = sendPeerChatMessage(md, request)
	case mtproto.TLConstructor_CRC32_inputPeerChannel:
		sentMessage, err = sendPeerChannelMessage(md, request)
	default:
		err = fmt.Errorf("MessagesSendMessage - Invalid inpuet peer {%v}", request.GetPeer())
		glog.Error(err)
		return nil, err
	}

	glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
	return sentMessage.To_Updates(), nil

	/*
		peer := base.FromInputPeer(request.GetPeer())
		message := mtproto.NewTLMessage()

		message.SetSilent(request.GetSilent())
		now := int32(time.Now().Unix())
		// TODO(@benqi): ???
		// request.Background
		// request.NoWebpage
		// request.ClearDraft
		message.SetFromId(md.UserId)
		if peer.PeerType == base.PEER_SELF {
			to := &mtproto.TLPeerUser{ Data2: &mtproto.Peer_Data{
				UserId: md.UserId,
			}}
			message.SetToId(to.To_Peer())
		} else {
			message.SetToId(peer.ToPeer())
		}
		message.SetMessage(request.Message)
		message.SetReplyToMsgId(request.ReplyToMsgId)
		message.SetReplyMarkup(request.ReplyMarkup)
		message.SetEntities(request.Entities)
		message.SetDate(now)
		// glog.Infof("metadata: {%v}, rpcMetaData: {%v}", md, rpcMetaData)
		sentMessage := mtproto.NewTLUpdateShortSentMessage()
		switch peer.PeerType {
		case base.PEER_SELF:
			// 1. SaveMessage
			messageId := model.GetMessageModel().CreateHistoryMessage2(md.UserId, peer, request.RandomId, now, message.To_Message())
			// 2. MessageBoxes
			pts := model.GetMessageModel().CreateMessageBoxes(md.UserId, message.GetFromId(), peer.PeerType, md.UserId, false, messageId)
			// 3. dialog
			model.GetDialogModel().CreateOrUpdateByLastMessage(md.UserId, peer.PeerType, md.UserId, messageId, message.GetMentioned(), false)
			// 推送给sync
			// 推给客户端的updates
			updates := mtproto.NewTLUpdateShortMessage()
			updates.SetId(int32(messageId))
			updates.SetUserId(md.UserId)
			// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
			updates.SetPts(pts)
			updates.SetPtsCount(1)
			updates.SetMessage(request.Message)
			updates.SetDate(now)
			delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
				md.AuthId,
				md.SessionId,
				md.NetlibSessionId,
				[]int32{md.UserId},
				updates.To_Updates().Encode())
			// 返回给客户端
			// sentMessage := &mtproto.TLUpdateShortSentMessage{}
			sentMessage.SetOut(true)
			sentMessage.SetId(int32(messageId))
			// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
			sentMessage.SetPts(pts)
			sentMessage.SetPtsCount(1)
			sentMessage.SetDate(now)
			sentMessage.SetMedia(mtproto.NewTLMessageMediaEmpty().To_MessageMedia())
			glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
			// reply = sentMessage.ToUpdates()
		case base.PEER_CHAT:
			// 返回给客户端
			// sentMessage := &mtproto.TLUpdateShortSentMessage{}
			sentMessage.SetOut(true)
			// sentMessage.Id = int32(messageId)
			// sentMessage.Pts = outPts
			sentMessage.SetPtsCount(1)
			sentMessage.SetDate(message.GetDate())
			// 1. SaveMessage
			messageId := model.GetMessageModel().CreateHistoryMessage2(md.UserId, peer, request.RandomId, now, message.To_Message())
			participants := model.GetChatModel().GetChatParticipants(peer.PeerId)
			var userId int32 = 0
			for _, participan := range participants.GetParticipants() {
				switch participan.GetConstructor() {
				case mtproto.TLConstructor_CRC32_chatParticipantCreator:
					userId = participan.GetData2().GetUserId()
				case mtproto.TLConstructor_CRC32_chatParticipantAdmin:
					userId = participan.GetData2().GetUserId()
				case mtproto.TLConstructor_CRC32_chatParticipant:
					userId = participan.GetData2().GetUserId()
				}
				// 2. MessageBoxes
				outgoing := userId == md.UserId
				pts := model.GetMessageModel().CreateMessageBoxes(userId, md.UserId, peer.PeerType, peer.PeerId, outgoing, messageId)
				model.GetDialogModel().CreateOrUpdateByLastMessage(userId, peer.PeerType, peer.PeerId, messageId, message.GetMentioned(), false)
				// inPts := model.GetMessageModel().CreateMessageBoxes(peer.PeerId, message.FromId, peer, true, messageId)
				// 3. dialog
				// model.GetDialogModel().CreateOrUpdateByLastMessage(peer.PeerId, peer, messageId, message.Mentioned)
				// 推送给sync
				// 推给客户端的updates
				updates := mtproto.NewTLUpdateShortChatMessage()
				updates.SetId(int32(messageId))
				updates.SetFromId(md.UserId)
				updates.SetChatId(peer.PeerId)
				// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
				updates.SetPts(pts)
				updates.SetPtsCount(1)
				updates.SetMessage(request.Message)
				updates.SetDate(message.GetDate())
				if md.UserId == userId {
					sentMessage.SetId(int32(messageId))
					// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
					sentMessage.SetPts(pts-1)
					delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
						md.AuthId,
						md.SessionId,
						md.NetlibSessionId,
						[]int32{userId},
						updates.To_Updates().Encode())
				} else {
					delivery.GetDeliveryInstance().DeliveryUpdates(
						md.AuthId,
						md.SessionId,
						md.NetlibSessionId,
						[]int32{userId},
						updates.To_Updates().Encode())
				}
			}
			glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
			// reply = sentMessage.ToUpdates()
		case base.PEER_USER:
			// 1. SaveMessage
			messageId := model.GetMessageModel().CreateHistoryMessage2(md.UserId, peer, request.RandomId, now, message.To_Message())
			// 2. MessageBoxes
			outPts := model.GetMessageModel().CreateMessageBoxes(md.UserId, message.GetFromId(), peer.PeerType, peer.PeerId, false, messageId)
			inPts := model.GetMessageModel().CreateMessageBoxes(peer.PeerId, message.GetFromId(), peer.PeerType, md.UserId, true, messageId)
			// 3. dialog
			model.GetDialogModel().CreateOrUpdateByLastMessage(md.UserId, peer.PeerType, peer.PeerId, messageId, message.GetMentioned(), false)
			model.GetDialogModel().CreateOrUpdateByLastMessage(peer.PeerId, peer.PeerType, md.UserId, messageId, message.GetMentioned(), true)
			// 推送给sync
			// 推给客户端的updates
			updates := mtproto.NewTLUpdateShortMessage()
			updates.SetId(int32(messageId))
			updates.SetUserId(md.UserId)
			// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
			updates.SetPts(inPts)
			updates.SetPtsCount(1)
			updates.SetMessage(request.Message)
			updates.SetDate(message.GetDate())
			delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
				md.AuthId,
				md.SessionId,
				md.NetlibSessionId,
				[]int32{md.UserId, peer.PeerId},
				updates.To_Updates().Encode())

	/ *
			// 用户在线订阅
			updateShort := mtproto.NewTLUpdateShort();
			updateUserStatus := mtproto.NewTLUpdateUserStatus()
			updateUserStatus.SetUserId(md.UserId)
			userStatus := mtproto.NewTLUserStatusOffline()
			userStatus.SetWasOnline(now)
			updateUserStatus.SetStatus(userStatus.To_UserStatus())
			updateShort.SetUpdate(updateUserStatus.To_Update())
			updateShort.SetDate(now)
			delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
				md.AuthId,
				md.SessionId,
				md.NetlibSessionId,
				[]int32{peer.PeerId},
				updateShort.To_Updates().Encode())
	 * /

			// 返回给客户端
			// sentMessage := &mtproto.TLUpdateShortSentMessage{}
			sentMessage.SetOut(true)
			sentMessage.SetId(int32(messageId))
			// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
			sentMessage.SetPts(outPts)
			sentMessage.SetPtsCount(1)
			sentMessage.SetDate(now)
			// glog.Infof("MessagesSendMessage - reply: %v", sentMessage)
			// reply = sentMessage.ToUpdates()
		case base.PEER_CHANNEL:
		default:
			panic(mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_BAD_REQUEST), "InputPeer invalid"))
		}

		glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
		return sentMessage.To_Updates(), nil
	 */
}

func makeMessageBySendMessage(request *mtproto.TLMessagesSendMessage) *mtproto.TLMessage {
	// peer := base.FromInputPeer(request.GetPeer())
	message := mtproto.NewTLMessage()

	message.SetSilent(request.GetSilent())
	//// TODO(@benqi): ???
	//// request.Background
	//// request.NoWebpage
	//// request.ClearDraft
	//message.SetFromId(md.UserId)
	//if peer.PeerType == base.PEER_SELF {
	//	to := &mtproto.TLPeerUser{ Data2: &mtproto.Peer_Data{
	//		UserId: md.UserId,
	//	}}
	//	message.SetToId(to.To_Peer())
	//} else {
	//	message.SetToId(peer.ToPeer())
	//}

	message.SetMessage(request.Message)
	message.SetReplyToMsgId(request.ReplyToMsgId)
	message.SetReplyMarkup(request.ReplyMarkup)
	message.SetEntities(request.Entities)
	message.SetDate(int32(time.Now().Unix()))

	return message
}

func sendPeerSelfMessage(md *zproto.RpcMetadata, request *mtproto.TLMessagesSendMessage) (*mtproto.TLUpdateShortSentMessage, error) {
	// message
	message := makeMessageBySendMessage(request)
	message.SetFromId(md.UserId)
	// peer
	peer := &base.PeerUtil{PeerType: base.PEER_USER, PeerId: md.UserId}
	message.SetToId(peer.ToPeer())

	// message
	ids := model.GetMessageModel().SendMessage(md.UserId, base.PEER_USER, md.UserId, request.GetRandomId(), message.To_Message())

	// model.GetUpdatesModel().AddPtsToUpdatesQueue(md.UserId, )
	// 推给客户端的updates
	updates := mtproto.NewTLUpdateShortMessage()
	// sentMessage.SetOut(true)
	updates.SetId(ids[0].MessageBoxId)
	updates.SetUserId(md.UserId)
	// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
	// updates.SetPts(pts)
	// updates.SetPtsCount(1)
	updates.SetMessage(request.Message)
	updates.SetDate(message.GetDate())
	state, err := sync.GetSyncClient().SyncUpdateShortMessage(md.AuthId, md.SessionId, md.NetlibSessionId, md.UserId, md.UserId, updates)
	if err != nil {
		return nil, err
	}

	sentMessage := mtproto.NewTLUpdateShortSentMessage()
	// sentMessage := &mtproto.TLUpdateShortSentMessage{}
	sentMessage.SetOut(true)
	sentMessage.SetId(int32(ids[0].MessageBoxId))
	// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
	sentMessage.SetPts(state.Pts)
	sentMessage.SetPtsCount(state.PtsCount)
	sentMessage.SetDate(message.GetDate())
	sentMessage.SetMedia(mtproto.NewTLMessageMediaEmpty().To_MessageMedia())

	////// 1. SaveMessage
	////messageId := model.GetMessageModel().CreateHistoryMessage2(md.UserId, peer, request.RandomId, message.GetDate(), message.To_Message())
	////// 2. MessageBoxes
	////pts := model.GetMessageModel().CreateMessageBoxes(md.UserId, message.GetFromId(), peer.PeerType, md.UserId, false, messageId)
	////// 3. dialog
	////model.GetDialogModel().CreateOrUpdateByLastMessage(md.UserId, peer.PeerType, md.UserId, messageId, message.GetMentioned(), false)
	//
	//// 推送给sync
	//pts := int32(model.GetSequenceModel().NextPtsId(base2.Int32ToString(md.UserId)))
	//model.GetUpdatesModel().AddPtsToUpdatesQueue(md.UserId, pts, base.PEER_USER, md.UserId, model.PTS_MESSAGE_INBOX, ids[0].MessageBoxId, 0)
	//
	//
	//delivery.GetDeliveryInstance().DeliveryUpdatesNotMe(
	//	md.AuthId,
	//	md.SessionId,
	//	md.NetlibSessionId,
	//	[]int32{md.UserId},
	//	updatesData)
	//// 返回给客户端
	glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
	// reply = sentMessage.ToUpdates()
	return sentMessage, nil
}

func sendPeerUserMessage(md *zproto.RpcMetadata, request *mtproto.TLMessagesSendMessage) (*mtproto.TLUpdateShortSentMessage, error) {
	// message
	message := makeMessageBySendMessage(request)
	message.SetFromId(md.UserId)
	// peer
	peer := &base.PeerUtil{PeerType: base.PEER_USER, PeerId: request.GetPeer().GetData2().GetUserId()}
	message.SetToId(peer.ToPeer())

	// message
	ids := model.GetMessageModel().SendMessage(md.UserId, base.PEER_USER, peer.PeerId, request.GetRandomId(), message.To_Message())


	// model.GetUpdatesModel().AddPtsToUpdatesQueue(md.UserId, )
	// 推给客户端的updates
	updates := mtproto.NewTLUpdateShortMessage()
	// sentMessage.SetOut(true)
	updates.SetId(ids[0].MessageBoxId)
	updates.SetUserId(md.UserId)
	// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
	// updates.SetPts(pts)
	// updates.SetPtsCount(1)
	updates.SetMessage(request.Message)
	updates.SetDate(message.GetDate())
	state, err := sync.GetSyncClient().SyncUpdateShortMessage(md.AuthId, md.SessionId, md.NetlibSessionId, md.UserId, md.UserId, updates)
	if err != nil {
		return nil, err
	}

	sentMessage := mtproto.NewTLUpdateShortSentMessage()
	sentMessage.SetOut(true)
	sentMessage.SetId(int32(ids[0].MessageBoxId))
	// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
	sentMessage.SetPts(state.Pts)
	sentMessage.SetPtsCount(state.PtsCount)
	sentMessage.SetDate(message.GetDate())
	sentMessage.SetMedia(mtproto.NewTLMessageMediaEmpty().To_MessageMedia())

	//// 1. SaveMessage
	//messageId := model.GetMessageModel().CreateHistoryMessage2(md.UserId, peer, request.RandomId, message.GetDate(), message.To_Message())
	//// 2. MessageBoxes
	//pts := model.GetMessageModel().CreateMessageBoxes(md.UserId, message.GetFromId(), peer.PeerType, md.UserId, false, messageId)
	//// 3. dialog
	//model.GetDialogModel().CreateOrUpdateByLastMessage(md.UserId, peer.PeerType, md.UserId, messageId, message.GetMentioned(), false)

	// var myPts int = 0
	for _, idPair := range ids {
		// model.GetUpdatesModel().AddPtsToUpdatesQueue(md.UserId, )
		// 推给客户端的updates
		updates2 := mtproto.NewTLUpdateShortMessage()
		updates2.SetId(idPair.MessageBoxId)
		updates2.SetUserId(md.UserId)
		// TODO(@benqi): 暂时这样实现验证发消息是否有问题，有问题的
		//updates.SetPts(pts)
		//updates.SetPtsCount(1)
		updates2.SetMessage(request.Message)
		updates2.SetDate(message.GetDate())
		sync.GetSyncClient().PushUpdateShortMessage(md.UserId, idPair.UserId, updates2)
	}

	glog.Infof("MessagesSendMessage - reply: %s", logger.JsonDebugData(sentMessage))
	// reply = sentMessage.ToUpdates()
	return sentMessage, nil
}

func sendPeerChatMessage(md *zproto.RpcMetadata, request *mtproto.TLMessagesSendMessage) (*mtproto.TLUpdateShortSentMessage, error) {
	return nil, nil
}

func sendPeerChannelMessage(md *zproto.RpcMetadata, request *mtproto.TLMessagesSendMessage) (*mtproto.TLUpdateShortSentMessage, error) {
	return nil, nil
}

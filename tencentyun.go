package ula

import (
	`fmt`
	`strconv`
	`strings`
	`time`

	`github.com/rs/xid`
	`github.com/storezhang/gox`
)

var _ executor = (*tencentyun)(nil)

type tencentyun struct{}

func (t *tencentyun) createLive(_ *CreateLiveReq, _ *options) (id string, err error) {
	// 取得和直播返回的直播编号
	id = xid.New().String()

	return
}

func (t *tencentyun) getPushUrls(id string, options *options) (urls []Url, err error) {
	urls = []Url{{
		Type: VideoFormatTypeRtmp,
		Link: t.makeUrl(
			VideoFormatTypeRtmp,
			options.tencentyun.push,
			id,
			1,
			options,
		),
	}}

	return
}

func (t *tencentyun) getPullCameras(id string, options *options) (cameras []Camera, err error) {
	cameras = []Camera{{
		Index: 1,
		Videos: []Video{{
			Type: VideoTypeOriginal,
			Urls: []Url{{
				Type: VideoFormatTypeRtmp,
				Link: t.makeUrl(
					VideoFormatTypeRtmp,
					options.tencentyun.pull,
					id,
					1,
					options,
				),
			}, {
				Type: VideoFormatTypeHls,
				Link: t.makeUrl(
					VideoFormatTypeHls,
					options.tencentyun.pull,
					id,
					1,
					options,
				),
			}, {
				Type: VideoFormatTypeFlv,
				Link: t.makeUrl(
					VideoFormatTypeFlv,
					options.tencentyun.pull,
					id,
					1,
					options,
				),
			}, {
				Type: VideoFormatTypeRtc,
				Link: t.makeUrl(
					VideoFormatTypeRtc,
					options.tencentyun.pull,
					id,
					1,
					options,
				),
			}},
		}},
	}}

	return
}

func (t *tencentyun) stop(_ string, _ *options) (success bool, err error) {
	success = true

	return
}

func (t *tencentyun) getViewerNum(id string, options *options) (viewerNum int64, err error) {
	return
}

func (t *tencentyun) makeUrl(
	formatType VideoFormatType,
	domain *domain,
	id string,
	camera int8,
	options *options,
) (url string) {
	expirationTime := time.Now().Add(options.expired).Unix()
	expirationHex := strings.ToUpper(strconv.FormatInt(expirationTime, 16))
	streamName := fmt.Sprintf("%s-%d", id, camera)
	key, _ := gox.Md5(fmt.Sprintf("%s%s%s", domain.key, streamName, expirationHex))

	switch formatType {
	case VideoFormatTypeRtmp:
		url = fmt.Sprintf(
			"rtmp://%s/live/%s?txSecret=%s&txTime=%s",
			domain.name,
			streamName,
			key,
			expirationHex,
		)
	case VideoFormatTypeRtc:
		url = fmt.Sprintf(
			"webrtc://%s/live/%s?txSecret=%s&txTime=%s",
			domain.name,
			streamName,
			key,
			expirationHex,
		)
	case VideoFormatTypeFlv:
		url = fmt.Sprintf(
			"%s://%s/live/%s.flv?txSecret=%s&txTime=%s",
			options.scheme,
			domain.name,
			streamName,
			key,
			expirationHex,
		)
	case VideoFormatTypeHls:
		url = fmt.Sprintf(
			"%s://%s/live/%s.m3u8?txSecret=%s&txTime=%s",
			options.scheme,
			domain.name,
			streamName,
			key,
			expirationHex,
		)
	default:
		url = fmt.Sprintf(
			"%s://%s/live/%s.flv?txSecret=%s&txTime=%s",
			options.scheme,
			domain.name,
			streamName,
			key,
			expirationHex,
		)
	}

	// 超低延时播放：支持400ms左右的超低延迟播放是腾讯云直播播放器的一个特点，它可以用于一些对时延要求极为苛刻的场景，例如远程夹娃娃或者主播连麦等
	// 播放地址需要带防盗链：播放URL不能用普通的CDN URL，必须要带防盗链签名和bizid参数，防盗链签名的计算方法请参见防盗链计算
	// 播放类型需要指定ACC：在调用startPlay函数时，需要指定type为PLAY_TYPE_LIVE_RTMP_ACC，SDK会使用RTMP-UDP协议拉取直播流
	// 该功能有并发播放限制：目前最多同时10路并发播放，避免因为盲目追求低延时而产生不必要的费用损失
	// OBS的延时是不达标的：推流端如果是TXLivePusher，请使用setVideoQuality将quality设置为MAIN_PUBLISHER或者VIDEO_CHAT
	// 该功能按播放时长收费：本功能按照播放时长收费，费用跟拉流的路数有关系，跟音视频流的码率无关，具体价格请参考 价格总览
	if 0 != options.tencentyun.bizId {
		url = fmt.Sprintf("%s&bizid=%d", url, options.tencentyun.bizId)
	}

	return
}

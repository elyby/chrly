package db

import (
	"bytes"
	"compress/zlib"
	"io"
	"strings"

	"github.com/valyala/fastjson"
)

type ProfileSerializer interface {
	Serialize(profile *Profile) ([]byte, error)
	Deserialize(value []byte) (*Profile, error)
}

func NewJsonSerializer() *JsonSerializer {
	return &JsonSerializer{
		parserPool: &fastjson.ParserPool{},
	}
}

type JsonSerializer struct {
	parserPool *fastjson.ParserPool
}

// Reasons for manual JSON serialization:
// 1. The Profile must be pure and must not contain tags.
// 2. Without tags it's impossible to apply omitempty during serialization.
// 3. Without omitempty we significantly inflate the storage size, which is critical for large deployments.
// Since the JSON structure in this case is very simple, it's very easy to write a manual serialization,
// achieving all constraints above.
func (s *JsonSerializer) Serialize(profile *Profile) ([]byte, error) {
	var builder strings.Builder
	// Prepare for the worst case (e.g. long username, long textures links, long Mojang textures and signature)
	// to prevent additional memory allocations during serialization
	builder.Grow(1536)
	builder.WriteString(`{"uuid":"`)
	builder.WriteString(profile.Uuid)
	builder.WriteString(`","username":"`)
	builder.WriteString(profile.Username)
	builder.WriteString(`"`)
	if profile.SkinUrl != "" {
		builder.WriteString(`,"skinUrl":"`)
		builder.WriteString(profile.SkinUrl)
		builder.WriteString(`"`)
		if profile.SkinModel != "" {
			builder.WriteString(`,"skinModel":"`)
			builder.WriteString(profile.SkinModel)
			builder.WriteString(`"`)
		}
	}

	if profile.CapeUrl != "" {
		builder.WriteString(`,"capeUrl":"`)
		builder.WriteString(profile.CapeUrl)
		builder.WriteString(`"`)
	}

	if profile.MojangTextures != "" {
		builder.WriteString(`,"mojangTextures":"`)
		builder.WriteString(profile.MojangTextures)
		builder.WriteString(`","mojangSignature":"`)
		builder.WriteString(profile.MojangSignature)
		builder.WriteString(`"`)
	}

	builder.WriteString("}")

	return []byte(builder.String()), nil
}

func (s *JsonSerializer) Deserialize(value []byte) (*Profile, error) {
	parser := s.parserPool.Get()
	defer s.parserPool.Put(parser)
	v, err := parser.ParseBytes(value)
	if err != nil {
		return nil, err
	}

	profile := &Profile{
		Uuid:            string(v.GetStringBytes("uuid")),
		Username:        string(v.GetStringBytes("username")),
		SkinUrl:         string(v.GetStringBytes("skinUrl")),
		SkinModel:       string(v.GetStringBytes("skinModel")),
		CapeUrl:         string(v.GetStringBytes("capeUrl")),
		MojangTextures:  string(v.GetStringBytes("mojangTextures")),
		MojangSignature: string(v.GetStringBytes("mojangSignature")),
	}

	return profile, nil
}

func NewZlibEncoder(serializer ProfileSerializer) *ZlibEncoder {
	return &ZlibEncoder{serializer}
}

type ZlibEncoder struct {
	serializer ProfileSerializer
}

func (s *ZlibEncoder) Serialize(profile *Profile) ([]byte, error) {
	serialized, err := s.serializer.Serialize(profile)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	writer := zlib.NewWriter(&buff)
	_, err = writer.Write(serialized)
	if err != nil {
		return nil, err
	}

	_ = writer.Close()

	return buff.Bytes(), nil
}

func (s *ZlibEncoder) Deserialize(value []byte) (*Profile, error) {
	buff := bytes.NewReader(value)
	reader, err := zlib.NewReader(buff)
	if err != nil {
		return nil, err
	}

	resultBuffer := new(bytes.Buffer)
	_, err = io.Copy(resultBuffer, reader)
	if err != nil {
		return nil, err
	}

	_ = reader.Close()

	return s.serializer.Deserialize(resultBuffer.Bytes())
}

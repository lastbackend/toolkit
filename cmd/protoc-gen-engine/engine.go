/*
Copyright [2014] - [2021] The Last.Backend authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"fmt"
	"strings"
)

const deprecationComment = "// Deprecated: Do not use."

const (
	contextPackage = protogen.GoImportPath("context")
	apiPackage     = protogen.GoImportPath("gitlab.com/lastbackend/engine/api")
	clientPackage  = protogen.GoImportPath("gitlab.com/lastbackend/engine/client")
	serverPackage  = protogen.GoImportPath("gitlab.com/lastbackend/engine/server")
	protoPackage   = protogen.GoImportPath("github.com/golang/protobuf/proto")
)

// generateFile generates a _engine.pb.go file containing gRPC service definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_engine.pb.go"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-engine. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-engine v", version)
	g.P("// - protoc            ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(gen, file, g)
	return g
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

// generateFileContent generates the gRPC service definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}

	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the grpc package it is being compiled against.")
	g.P("const _ = ", protoPackage.Ident("ProtoPackageIsVersion3")) // When changing, update version number above.
	g.P("// Reference imports to suppress errors if they are not otherwise used.")
	//g.P("var _ ", apiPackage, ".Endpoint")
	g.P("var _ ", contextPackage.Ident("Context"))
	g.P("var _ ", apiPackage.Ident("Endpoint"))
	g.P("var _ ", clientPackage.Ident("Option"))
	g.P("var _ ", serverPackage.Ident("Option"))
	g.P()
	for _, service := range file.Services {
		genService(gen, file, g, service)
	}
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {

	var (
		clientType = service.GoName + "Service"
		serverType = service.GoName + "Handler"
	)

	g.P("// Api Endpoints for ", service.GoName, " service.")
	g.P("func New", service.GoName, "Endpoints () []*", apiPackage.Ident("Endpoint"), " {")
	g.P("return []*", apiPackage.Ident("Endpoint"), " {")
	for _, method := range service.Methods {
		options, ok := method.Desc.Options().(*descriptorpb.MethodOptions)
		if ok && options != nil && proto.HasExtension(options, annotations.E_Http) {
			g.P("&", apiPackage.Ident("Endpoint"), " {")
			generateEndpoint(g, method)
			g.P("},")
		}
	}
	g.P("}")
	g.P("}")
	g.P()

	g.P("// ", clientType, " is the client API for ", service.GoName, " service.")
	g.P("//")
	g.P("// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.")

	// Client interface.
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	g.Annotate(clientType, service.Location)
	g.P("type ", clientType, " interface {")
	for _, method := range service.Methods {
		g.Annotate(clientType+"."+method.GoName, method.Location)
		if method.Desc.Options().(*descriptorpb.MethodOptions).GetDeprecated() {
			g.P(deprecationComment)
		}
		g.P(method.Comments.Leading, clientSignature(g, method))
	}
	g.P("}")
	g.P()

	// Client structure.
	g.P("type ", unexport(clientType), " struct {")
	g.P("c ", clientPackage.Ident("Client"))
	g.P("name string")
	g.P("}")
	g.P()

	// NewClient factory.
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P(deprecationComment)
	}

	// NewClient factory.
	g.P("func New", clientType, " (name string, c ", clientPackage.Ident("Client"), ") ", clientType, " {")
	g.P("return &", unexport(clientType), "{")
	g.P("c: c,")
	g.P("name: name,")
	g.P("}")
	g.P("}")
	g.P()

	// Client method implementations.
	var methodIndex, streamIndex int
	for _, method := range service.Methods {
		if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
			// Unary RPC method
			genClientMethod(gen, file, g, method, methodIndex)
			methodIndex++
		} else {
			// Streaming RPC method
			genClientMethod(gen, file, g, method, streamIndex)
			streamIndex++
		}
	}

	g.P("// Server API for ", service.GoName, " service")
	g.P("// ", serverType, " is the server API for ", service.GoName, " service.")

	// Server interface.
	g.P("type ", serverType, " interface {")
	for _, method := range service.Methods {
		g.P(serverSignature(g, method))
	}
	g.P("register(s ", serverPackage.Ident("Server"), ") error")
	g.P("}")
	g.P()

	// Server registration.
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P(deprecationComment)
	}
	g.Annotate(serverType, service.Location)
	g.P("func (h *", unexport(serverType), ") register(s ", serverPackage.Ident("Server"), ") error {")
	g.Annotate(serverType, service.Location)
	g.P("var opts = make([]", serverPackage.Ident("HandlerOption"), ",0)")
	g.P("type ", unexport(service.GoName), " interface {")
	for _, method := range service.Methods {
		g.Annotate(serverType+"."+method.GoName, method.Location)
		if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
			inType := g.QualifiedGoIdent(method.Input.GoIdent)
			outType := g.QualifiedGoIdent(method.Output.GoIdent)
			g.P(method.Desc.Name(), "(ctx ", contextPackage.Ident("Context"), ", in *", inType, ") (*", outType, ", error)")
			continue
		}
		g.P(method.Desc.Name(), "(ctx ", contextPackage.Ident("Context"), ", stream server.Stream) error")
	}
	g.P("register(s ", serverPackage.Ident("Server"), ") error")
	g.P("}")
	g.P("type ", service.GoName, " struct {")
	g.P(unexport(service.GoName))
	g.P("}")
	for _, method := range service.Methods {
		options, ok := method.Desc.Options().(*descriptorpb.MethodOptions)
		if ok && options != nil && proto.HasExtension(options, annotations.E_Http) {
			g.P("opts = append(opts, ", apiPackage.Ident("WithEndpoint"), "(&", apiPackage.Ident("Endpoint"), "{")
			generateEndpoint(g, method)
			g.P("}))")
		}
	}
	g.P("return s.Handle(s.NewHandler(&", service.GoName, "{h}, opts...))")
	g.P("}")
	g.P()

	g.P("type ", unexport(service.GoName), "Handler struct {")
	g.P(serverType)
	g.P("}")

	// Server handler implementations.
	var handlerNames []string
	for _, method := range service.Methods {
		hname := genServerMethod(g, method)
		handlerNames = append(handlerNames, hname)
	}

}

// generateEndpoint creates the api endpoint
func generateEndpoint(g *protogen.GeneratedFile, method *protogen.Method) {

	var (
		service = method.Parent
	)

	options, ok := method.Desc.Options().(*descriptorpb.MethodOptions)
	if !ok || options == nil || !proto.HasExtension(options, annotations.E_Http) {
		return
	}

	// http rules
	httpRule, ok := proto.GetExtension(options, annotations.E_Http).(*annotations.HttpRule)
	if !ok {
		return
	}

	var meth string
	var path string
	switch {
	case len(httpRule.GetDelete()) > 0:
		meth = "DELETE"
		path = httpRule.GetDelete()
	case len(httpRule.GetGet()) > 0:
		meth = "GET"
		path = httpRule.GetGet()
	case len(httpRule.GetPatch()) > 0:
		meth = "PATCH"
		path = httpRule.GetPatch()
	case len(httpRule.GetPost()) > 0:
		meth = "POST"
		path = httpRule.GetPost()
	case len(httpRule.GetPut()) > 0:
		meth = "PUT"
		path = httpRule.GetPut()
	}
	if len(meth) == 0 || len(path) == 0 {
		return
	}

	g.P("Name:", fmt.Sprintf(`"%s.%s",`, service.GoName, method.GoName))
	g.P("Path:", fmt.Sprintf(`[]string{"%s"},`, path))
	g.P("Method:", fmt.Sprintf(`[]string{"%s"},`, meth))
	if len(httpRule.GetGet()) == 0 {
		g.P("Body:", fmt.Sprintf(`"%s",`, httpRule.GetBody()))
	}
	if method.Desc.IsStreamingServer() || method.Desc.IsStreamingClient() {
		g.P("Stream: true,")
	}
	g.P(`Handler: "rpc",`)
}

func clientSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	s := method.GoName + "(ctx " + g.QualifiedGoIdent(contextPackage.Ident("Context"))
	if !method.Desc.IsStreamingClient() {
		s += ", in *" + g.QualifiedGoIdent(method.Input.GoIdent)
	}
	s += ", opts ..." + g.QualifiedGoIdent(clientPackage.Ident("CallOption")) + ") ("
	if !method.Desc.IsStreamingClient() && !method.Desc.IsStreamingServer() {
		s += "*" + g.QualifiedGoIdent(method.Output.GoIdent)
	} else {
		s += method.Parent.GoName + "_" + method.GoName + "Service"
	}
	s += ", error)"
	return s
}

func genClientMethod(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, method *protogen.Method, index int) {

	var (
		service     = method.Parent
		serviceName = method.Parent.GoName + "Service"
		methodName  = method.GoName
		inType      = g.QualifiedGoIdent(method.Input.GoIdent)
		outType     = g.QualifiedGoIdent(method.Output.GoIdent)
	)

	reqMethod := fmt.Sprintf("%s.%s", service.GoName, methodName)

	g.P("func (c *", unexport(serviceName), ") ", clientSignature(g, method), "{")
	if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
		g.P(`req := c.c.NewRequest(c.name, "`, reqMethod, `", in)`)
		g.P("out := new(", outType, ")")
		g.P("err := ", `c.c.Call(ctx, req, out, opts...)`)
		g.P("if err != nil { return nil, err }")
		g.P("return out, nil")
		g.P("}")
		g.P()
		return
	}
	streamType := unexport(serviceName) + methodName
	g.P(`req := c.c.NewRequest(c.name, "`, reqMethod, `", &`, inType, `{})`)
	g.P("stream, err := c.c.Stream(ctx, req, opts...)")
	g.P("if err != nil { return nil, err }")

	if !method.Desc.IsStreamingClient() {
		g.P("if err := stream.Send(in); err != nil { return nil, err }")
	}

	g.P("return &", streamType, "{stream}, nil")
	g.P("}")
	g.P()

	genSend := method.Desc.IsStreamingClient()
	genRecv := method.Desc.IsStreamingServer()

	// Stream auxiliary types and methods.
	g.P("type ", service.GoName, "_", methodName, "Service interface {")
	g.P("Context() context.Context")
	g.P("SendMsg(interface{}) error")
	g.P("RecvMsg(interface{}) error")

	if genSend && !genRecv {
		// client streaming, the server will send a response upon close
		g.P("CloseAndRecv() (*", outType, ", error)")
	} else {
		g.P("Close() error")
	}

	if genSend {
		g.P("Send(*", inType, ") error")
	}
	if genRecv {
		g.P("Recv() (*", outType, ", error)")
	}
	g.P("}")
	g.P()

	g.P("type ", streamType, " struct {")
	g.P("stream ", clientPackage.Ident("Stream"))
	g.P("}")
	g.P()

	if genSend && !genRecv {
		// client streaming, the server will send a response upon close
		g.P("func (h *", streamType, ") CloseAndRecv() (*", outType, ", error) {")
		g.P("if err := h.stream.Close(); err != nil {")
		g.P("return nil, err")
		g.P("}")
		g.P("r := new(", outType, ")")
		g.P("err := h.RecvMsg(r)")
		g.P("return r, err")
		g.P("}")
		g.P()
	} else {
		g.P("func (h *", streamType, ") Close() error {")
		g.P("return h.stream.Close()")
		g.P("}")
		g.P()
	}

	g.P("func (h *", streamType, ") Context() context.Context {")
	g.P("return h.stream.Context()")
	g.P("}")
	g.P()

	g.P("func (h *", streamType, ") SendMsg(m interface{}) error {")
	g.P("return h.stream.Send(m)")
	g.P("}")
	g.P()

	g.P("func (h *", streamType, ") RecvMsg(m interface{}) error {")
	g.P("return h.stream.Recv(m)")
	g.P("}")
	g.P()

	if genSend {
		g.P("func (h *", streamType, ") Send(m *", inType, ") error {")
		g.P("return h.stream.Send(m)")
		g.P("}")
		g.P()
	}

	if genRecv {
		g.P("func (h *", streamType, ") Recv() (*", outType, ", error) {")
		g.P("m := new(", outType, ")")
		g.P("err := h.stream.Recv(m)")
		g.P("if err != nil {")
		g.P("return nil, err")
		g.P("}")
		g.P("return m, nil")
		g.P("}")
		g.P()
	}

}

func serverSignature(g *protogen.GeneratedFile, method *protogen.Method) string {
	var (
		reqArgs     []string
		ret         = "error"
		methodName  = method.GoName
		serviceName = method.Parent
	)

	reqArgs = append(reqArgs, g.QualifiedGoIdent(contextPackage.Ident("Context")))

	if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
		outType := g.QualifiedGoIdent(method.Output.GoIdent)
		ret = "(*" + outType + ", error)"
	}
	if !method.Desc.IsStreamingClient() {
		inType := g.QualifiedGoIdent(method.Input.GoIdent)
		reqArgs = append(reqArgs, "*"+inType)
	}
	if method.Desc.IsStreamingServer() || method.Desc.IsStreamingClient() {
		reqArgs = append(reqArgs, serviceName.GoName+"_"+methodName+"Stream")
	}

	return methodName + "(" + strings.Join(reqArgs, ", ") + ") " + ret
}

func genServerMethod(g *protogen.GeneratedFile, method *protogen.Method) string {

	var (
		serviceType = method.Parent.GoName + "Handler"
		serviceName = method.Parent.GoName
		methodName  = method.GoName
		inType      = g.QualifiedGoIdent(method.Input.GoIdent)
		outType     = g.QualifiedGoIdent(method.Output.GoIdent)
	)

	hname := fmt.Sprintf("_%s_%s_Handler", serviceName, method.GoName)

	if !method.Desc.IsStreamingServer() && !method.Desc.IsStreamingClient() {
		g.P("func (h *", unexport(serviceName), "Handler) ", methodName, "(ctx ", contextPackage.Ident("Context"), ", in *", inType, ") (*", outType, ", error) {")
		g.P("return h.", serviceType, ".", methodName, "(ctx, in)")
		g.P("}")
		g.P()
		return hname
	}
	streamType := unexport(serviceName) + methodName + "Stream"
	g.P("func (h *", unexport(serviceName), "Handler) ", methodName, "(ctx ", contextPackage.Ident("Context"), ", stream server.Stream) error {")
	if !method.Desc.IsStreamingClient() {
		g.P("m := new(", inType, ")")
		g.P("if err := stream.Recv(m); err != nil { return err }")
		g.P("return h.", serviceType, ".", methodName, "(ctx, m, &", streamType, "{stream})")
	} else {
		g.P("return h.", serviceType, ".", methodName, "(ctx, &", streamType, "{stream})")
	}
	g.P("}")
	g.P()

	genSend := method.Desc.IsStreamingServer()
	genRecv := method.Desc.IsStreamingClient()

	// Stream auxiliary types and methods.
	g.P("type ", serviceName, "_", methodName, "Stream interface {")
	g.P("Context() context.Context")
	g.P("SendMsg(interface{}) error")
	g.P("RecvMsg(interface{}) error")

	if !genSend {
		// client streaming, the server will send a response upon close
		g.P("SendAndClose(*", outType, ")  error")
	} else {
		g.P("Close() error")
	}

	if genSend {
		g.P("Send(*", outType, ") error")
	}

	if genRecv {
		g.P("Recv() (*", inType, ", error)")
	}

	g.P("}")
	g.P()

	g.P("type ", streamType, " struct {")
	g.P("stream ", serverPackage.Ident("Stream"))
	g.P("}")
	g.P()

	if !genSend {
		// client streaming, the server will send a response upon close
		g.P("func (h *", streamType, ") SendAndClose(in *", outType, ") error {")
		g.P("if err := h.SendMsg(in); err != nil {")
		g.P("return err")
		g.P("}")
		g.P("return h.stream.Close()")
		g.P("}")
		g.P()
	} else {
		// other types of rpc don't send a response when the stream closes
		g.P("func (h *", streamType, ") Close() error {")
		g.P("return h.stream.Close()")
		g.P("}")
		g.P()
	}

	g.P("func (h *", streamType, ") Context() context.Context {")
	g.P("return h.stream.Context()")
	g.P("}")
	g.P()

	g.P("func (h *", streamType, ") SendMsg(m interface{}) error {")
	g.P("return h.stream.Send(m)")
	g.P("}")
	g.P()

	g.P("func (h *", streamType, ") RecvMsg(m interface{}) error {")
	g.P("return h.stream.Recv(m)")
	g.P("}")
	g.P()

	if genSend {
		g.P("func (h *", streamType, ") Send(m *", outType, ") error {")
		g.P("return h.stream.Send(m)")
		g.P("}")
		g.P()
	}

	if genRecv {
		g.P("func (h *", streamType, ") Recv() (*", inType, ", error) {")
		g.P("m := new(", inType, ")")
		g.P("if err := h.stream.Recv(m); err != nil { return nil, err }")
		g.P("return m, nil")
		g.P("}")
		g.P()
	}

	return hname
}

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }

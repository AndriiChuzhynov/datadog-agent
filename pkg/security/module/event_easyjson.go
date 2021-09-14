// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package module

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjsonF642ad3eDecodeGithubComDataDogDatadogAgentPkgSecurityModule(in *jlexer.Lexer, out *Signal) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	out.AgentContext = new(AgentContext)
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "agent":
			if in.IsNull() {
				in.Skip()
				out.AgentContext = nil
			} else {
				if out.AgentContext == nil {
					out.AgentContext = new(AgentContext)
				}
				(*out.AgentContext).UnmarshalEasyJSON(in)
			}
		case "title":
			out.Title = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF642ad3eEncodeGithubComDataDogDatadogAgentPkgSecurityModule(out *jwriter.Writer, in Signal) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"agent\":"
		out.RawString(prefix[1:])
		if in.AgentContext == nil {
			out.RawString("null")
		} else {
			(*in.AgentContext).MarshalEasyJSON(out)
		}
	}
	{
		const prefix string = ",\"title\":"
		out.RawString(prefix)
		out.String(string(in.Title))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Signal) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF642ad3eEncodeGithubComDataDogDatadogAgentPkgSecurityModule(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Signal) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF642ad3eEncodeGithubComDataDogDatadogAgentPkgSecurityModule(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Signal) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF642ad3eDecodeGithubComDataDogDatadogAgentPkgSecurityModule(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Signal) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF642ad3eDecodeGithubComDataDogDatadogAgentPkgSecurityModule(l, v)
}
func easyjsonF642ad3eDecodeGithubComDataDogDatadogAgentPkgSecurityModule1(in *jlexer.Lexer, out *AgentContext) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "rule_id":
			out.RuleID = string(in.String())
		case "rule_version":
			out.RuleVersion = string(in.String())
		case "policy_name":
			out.PolicyName = string(in.String())
		case "policy_version":
			out.PolicyVersion = string(in.String())
		case "profile_name":
			out.ProfileName = string(in.String())
		case "profile_version":
			out.ProfileVersion = string(in.String())
		case "version":
			out.Version = string(in.String())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjsonF642ad3eEncodeGithubComDataDogDatadogAgentPkgSecurityModule1(out *jwriter.Writer, in AgentContext) {
	out.RawByte('{')
	first := true
	_ = first
	if in.RuleID != "" {
		const prefix string = ",\"rule_id\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.RuleID))
	}
	if in.RuleVersion != "" {
		const prefix string = ",\"rule_version\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.RuleVersion))
	}
	if in.PolicyName != "" {
		const prefix string = ",\"policy_name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.PolicyName))
	}
	if in.PolicyVersion != "" {
		const prefix string = ",\"policy_version\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.PolicyVersion))
	}
	if in.ProfileName != "" {
		const prefix string = ",\"profile_name\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.ProfileName))
	}
	if in.ProfileVersion != "" {
		const prefix string = ",\"profile_version\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.ProfileVersion))
	}
	if in.Version != "" {
		const prefix string = ",\"version\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Version))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v AgentContext) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonF642ad3eEncodeGithubComDataDogDatadogAgentPkgSecurityModule1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v AgentContext) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonF642ad3eEncodeGithubComDataDogDatadogAgentPkgSecurityModule1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *AgentContext) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonF642ad3eDecodeGithubComDataDogDatadogAgentPkgSecurityModule1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *AgentContext) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonF642ad3eDecodeGithubComDataDogDatadogAgentPkgSecurityModule1(l, v)
}

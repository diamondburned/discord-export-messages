package flags

import (
	"fmt"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

type StringEnum struct {
	value  string
	values []string
}

func NewStringEnum(values ...string) *StringEnum {
	return &StringEnum{
		value:  values[0],
		values: values,
	}
}

func (f *StringEnum) Set(s string) error {
	for _, v := range f.values {
		if s == v {
			f.value = s
			return nil
		}
	}
	return fmt.Errorf("invalid value: %s", s)
}

func (f *StringEnum) String() string {
	return f.value
}

func (f *StringEnum) Options() []string {
	return f.values
}

func (f *StringEnum) OptionsString() string {
	return strings.Join(f.values, ", ")
}

type Snowflake[T ~uint64] struct {
	v T
}

func InvalidSnowflake[T ~uint64]() Snowflake[T] {
	return Snowflake[T]{}
}

var (
	_ = Snowflake[discord.Snowflake]{}
	_ = Snowflake[discord.MessageID]{}
	_ = Snowflake[discord.ChannelID]{}
	_ = Snowflake[discord.GuildID]{}
)

func (f *Snowflake[T]) Set(s string) error {
	id, err := discord.ParseSnowflake(s)
	if err != nil {
		return err
	}
	f.v = T(id)
	return nil
}

func (f *Snowflake[T]) String() string {
	return discord.Snowflake(f.v).String()
}

func (f *Snowflake[T]) Value() T {
	return f.v
}

func (f *Snowflake[T]) IsValid() bool {
	return discord.Snowflake(f.v).IsValid()
}

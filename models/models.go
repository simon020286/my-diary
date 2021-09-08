package models

import (
	"fmt"

	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/emirpasic/gods/maps/treemap"
)

type AddLineOptions struct {
	DefaultValue string
	Order        bool
}

type Section struct {
	Name       string
	Value      string
	PrintTitle bool
}

type Content struct {
	Filename string
	values   *treemap.Map
}

func (c *Content) getValues() *treemap.Map {
	if c.values == nil {
		c.values = treemap.NewWithStringComparator()
	}

	return c.values
}

func (c *Content) Keys() []string {
	keys := c.getValues().Keys()
	s := make([]string, len(keys))
	for i, v := range keys {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func (c *Content) getValue(key string) (*arraylist.List, bool) {
	value, ok := c.getValues().Get(key)
	if !ok {
		return nil, ok
	}
	return value.(*arraylist.List), ok
}

func (c *Content) AddKey(key string) bool {
	_, ok := c.getValue(key)
	if ok {
		return false
	}

	c.getValues().Put(key, arraylist.New())
	return true
}

func (c *Content) RemoveKey(key string) {
	c.getValues().Remove(key)
}

func (c *Content) RemoveSection(key, sectionName string) {
	value, ok := c.getValue(key)
	if ok {
		index, _ := c.getSection(key, sectionName)
		value.Remove(index)
	}
}

func (c *Content) AddSection(key string, sectionName string) (*Section, bool) {
	existsValue, ok := c.getValue(key)
	if !ok {
		existsValue = arraylist.New()
		c.getValues().Put(key, existsValue)
	}
	for _, v := range existsValue.Values() {
		if v.(*Section).Name == sectionName {
			return v.(*Section), false
		}
	}
	newSection := &Section{
		Name: sectionName,
	}
	existsValue.Add(newSection)
	return newSection, true
}

func (c *Content) GetSectionNames(key string) ([]string, bool) {
	names := make([]string, 0)
	sections, ok := c.getValue(key)
	if !ok {
		return nil, ok
	}
	for _, v := range sections.Values() {
		names = append(names, v.(*Section).Name)
	}

	return names, true
}

func (c *Content) getSection(key string, section string) (int, *Section) {
	value, ok := c.getValue(key)
	if !ok {
		return -1, nil
	}

	index, finded := value.Find(func(index int, value interface{}) bool {
		current := value.(*Section)
		return current.Name == section
	})

	if index >= 0 {
		return index, finded.(*Section)
	}

	return index, nil
}

func (c *Content) GetSection(key string, section string) (*Section, bool) {
	index, finded := c.getSection(key, section)
	if index >= 0 {
		return finded, true
	}

	return nil, false
}

func (c *Content) AppendLineToSection(key string, section string, value string) {
	existsSection, ok := c.GetSection(key, section)
	if !ok {
		existsSection, _ = c.AddSection(key, section)
	} else {
		existsSection.Value += "\n"
	}

	existsSection.Value += value
}

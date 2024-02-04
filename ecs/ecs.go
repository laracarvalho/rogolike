package ecs

import (
	"strconv"
	"sync"
	"sync/atomic"
)

type EntityID uint32

func (id EntityID) String() string {
	return strconv.Itoa(int(id))
}

type ComponentID uint32

type Tag struct {
	flags   uint64 // limited to 64 components
	inverse bool
}

func (tag Tag) matches(smallertag Tag) bool {
	res := tag.flags&smallertag.flags == smallertag.flags

	if smallertag.inverse {
		return !res
	}

	return res
}

func (tag *Tag) binaryORInPlace(othertag Tag) *Tag {
	tag.flags |= othertag.flags
	return tag
}

func (tag *Tag) binaryNOTInPlace(othertag Tag) *Tag {
	tag.flags ^= othertag.flags
	return tag
}

func (tag Tag) clone() Tag {
	return tag
}

func (tag Tag) Inverse(values ...bool) Tag {

	clone := tag.clone()
	inverse := true
	if len(values) > 0 {
		inverse = values[0]
	}

	clone.inverse = inverse
	return clone
}

type View struct {
	tag      Tag
	entities QueryResultCollection
	lock     *sync.RWMutex
}

type QueryResultCollection []*QueryResult

func (coll QueryResultCollection) Entities() []*Entity {
	res := make([]*Entity, len(coll))
	for i, qr := range coll {
		res[i] = qr.Entity
	}

	return res
}

func (v *View) Get() QueryResultCollection {
	v.lock.RLock()
	defer v.lock.RUnlock()
	return v.entities
}

func (v *View) add(entity *Entity) {
	v.lock.Lock()
	v.entities = append(v.entities, entity.engine.GetEntityByID(
		entity.ID,
		v.tag,
	))
	v.lock.Unlock()
}

func (v *View) remove(entity *Entity) {
	v.lock.RLock()
	for i, qr := range v.entities {
		if qr.Entity.ID == entity.ID {
			maxbound := len(v.entities) - 1
			v.entities[maxbound], v.entities[i] = v.entities[i], v.entities[maxbound]
			v.entities = v.entities[:maxbound]
			break
		}
	}
	v.lock.RUnlock()
}

type Engine struct {
	lock            *sync.RWMutex
	entityIdInc     uint32
	componentNumInc uint32 // limited to 64

	entities     []*Entity
	entitiesByID map[EntityID]*Entity
	components   []*Component
	views        []*View
}

type Component struct {
	id         ComponentID
	tag        Tag
	datalock   *sync.RWMutex
	data       map[EntityID]interface{}
	destructor func(entity *Entity, data interface{})
}

func (component *Component) SetDestructor(destructor func(entity *Entity, data interface{})) {
	component.destructor = destructor
}

func (component *Component) GetID() ComponentID {
	return component.id
}

func (engine *Engine) CreateView(tagelements ...interface{}) *View {

	tag := BuildTag(tagelements...)

	view := &View{
		tag:  tag,
		lock: &sync.RWMutex{},
	}

	entities := engine.Query(tag)
	view.entities = make(QueryResultCollection, len(entities))
	engine.lock.Lock()
	for i, entityresult := range entities {
		view.entities[i] = entityresult
	}
	engine.views = append(engine.views, view)
	engine.lock.Unlock()

	return view
}

type Entity struct {
	ID     EntityID
	tag    Tag
	engine *Engine
}

func (entity *Entity) GetID() EntityID {
	return entity.ID
}

func NewEngine() *Engine {
	return &Engine{
		entityIdInc:     0,
		componentNumInc: 0,
		entitiesByID:    make(map[EntityID]*Entity),
		lock:            &sync.RWMutex{},
		views:           make([]*View, 0),
	}
}

func BuildTag(elements ...interface{}) Tag {

	tag := Tag{}

	for _, element := range elements {
		switch typedelement := element.(type) {
		case *Component:
			{
				tag.binaryORInPlace(typedelement.tag)
			}
		case Tag:
			{
				tag.binaryORInPlace(typedelement)
			}
		default:
			{
				panic("Invalid type passed to BuildTag; accepts only <*Component> and <Tag> types.")
			}
		}
	}

	return tag
}

func (engine *Engine) NewEntity() *Entity {

	nextId := ComponentID(atomic.AddUint32(&engine.componentNumInc, 1))
	id := nextId - 1 // to start at 0

	entity := &Entity{
		ID:     EntityID(id),
		engine: engine,
	}

	engine.lock.Lock()
	engine.entities = append(engine.entities, entity)
	engine.entitiesByID[entity.ID] = entity
	engine.lock.Unlock()

	return entity
}

func (engine *Engine) NewComponent() *Component {

	if engine.componentNumInc >= 63 {
		panic("Component overflow (limited to 64)")
	}

	nextId := ComponentID(atomic.AddUint32(&engine.componentNumInc, 1))
	id := nextId - 1 // to start at 0

	tag := Tag{
		flags:   (1 << id), // set bit on position corresponding to component number
		inverse: false,
	}

	component := &Component{
		id:       id,
		tag:      tag,
		data:     make(map[EntityID]interface{}),
		datalock: &sync.RWMutex{},
	}

	engine.lock.Lock()
	engine.components = append(engine.components, component)
	engine.lock.Unlock()

	return component
}

func (engine *Engine) GetEntityByID(id EntityID, tagelements ...interface{}) *QueryResult {

	engine.lock.RLock()
	res, ok := engine.entitiesByID[id]

	if !ok {
		engine.lock.RUnlock()
		return nil
	}

	tag := BuildTag(tagelements...)

	components := engine.fetchComponentsForEntity(res, tag)
	engine.lock.RUnlock()

	if components == nil {
		return nil
	}

	return &QueryResult{
		Entity:     res,
		Components: components,
	}

}

func (entity Entity) Matches(tag Tag) bool {
	return entity.tag.matches(tag)
}

func (entity *Entity) AddComponent(component *Component, componentdata interface{}) *Entity {
	component.datalock.Lock()
	component.data[entity.ID] = componentdata
	component.datalock.Unlock()

	component.datalock.RLock()

	tagbefore := entity.tag
	entity.tag.binaryORInPlace(component.tag)

	for _, view := range entity.engine.views {

		if !tagbefore.matches(view.tag) && entity.tag.matches(view.tag) {
			view.add(entity)
		}
	}

	component.datalock.RUnlock()
	return entity
}

func (entity *Entity) RemoveComponent(component *Component) *Entity {
	if component.destructor != nil {
		if data, ok := component.data[entity.ID]; ok {
			component.destructor(entity, data)
		}
	}

	component.datalock.Lock()
	delete(component.data, entity.ID)
	tagbefore := entity.tag
	entity.tag.binaryNOTInPlace(component.tag)

	for _, view := range entity.engine.views {
		if tagbefore.matches(view.tag) && !entity.tag.matches(view.tag) {
			view.remove(entity)
		}
	}

	component.datalock.Unlock()
	return entity
}

func (entity Entity) HasComponent(component *Component) bool {
	return entity.tag.matches(component.tag)
}

func (entity Entity) GetComponentData(component *Component) (interface{}, bool) {
	component.datalock.RLock()
	data, ok := component.data[entity.ID]
	component.datalock.RUnlock()

	return data, ok
}

func (engine *Engine) DisposeEntities(entities ...*Entity) {
	for _, entity := range entities {
		engine.DisposeEntity(entity)
	}
}

func (engine *Engine) DisposeEntity(entity interface{}) {

	var typedentity *Entity

	switch typeditem := entity.(type) {
	case *QueryResult:
		{
			typedentity = typeditem.Entity
		}
	case QueryResult:
		{
			typedentity = typeditem.Entity
		}
	case *Entity:
		{
			typedentity = typeditem
		}
	default:
		{
			panic("Invalid type passed to DisposeEntity; accepts only <*QueryResult>, <QueryResult> and <*Entity> types.")
		}
	}

	if typedentity == nil {
		return
	}

	engine.lock.Lock()
	for _, component := range engine.components {
		if typedentity.HasComponent(component) {
			typedentity.RemoveComponent(component)
		}
	}
	delete(engine.entitiesByID, typedentity.ID)
	engine.lock.Unlock()
}

type QueryResult struct {
	Entity     *Entity
	Components map[*Component]interface{}
}

func (engine *Engine) fetchComponentsForEntity(entity *Entity, tag Tag) map[*Component]interface{} {

	if !entity.tag.matches(tag) {
		return nil
	}

	componentMap := make(map[*Component]interface{})

	for _, component := range engine.components {
		if tag.matches(component.tag) {
			data, ok := entity.GetComponentData(component)
			if !ok {
				return nil // if one of the required components is not set, return nothing !
			}

			componentMap[component] = data
		}
	}

	return componentMap
}

func (engine *Engine) Query(tag Tag) QueryResultCollection {

	matches := make(QueryResultCollection, 0)

	engine.lock.RLock()
	for _, entity := range engine.entities {
		if entity.tag.matches(tag) {

			componentMap := make(map[*Component]interface{})

			for _, component := range engine.components {
				if tag.matches(component.tag) {
					data, _ := entity.GetComponentData(component)
					componentMap[component] = data
				}
			}

			matches = append(matches, &QueryResult{
				Entity:     entity,
				Components: componentMap,
			})

		}
	}
	engine.lock.RUnlock()

	return matches
}

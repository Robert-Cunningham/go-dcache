package dcache2

import (
	"sync"
	"fmt"
)

type Kmem_cache struct {
	Items []Dentry //map[Dentry]bool
	Lock sync.Mutex
}

func (kc *Kmem_cache) Kmem_cache_insert(d *Dentry) {
	kc.Lock.Lock()
	kc.Items = append(kc.Items, *d)
	//kmem_cache.Items[*d] = true
	kc.Lock.Unlock()
}

func (kmem_cache *Kmem_cache) Kmem_cache_free(d *Dentry) {
	kmem_cache.Lock.Lock()
	fmt.Println("Deleting from kmem_cache")
	//TODO
	//delete(kmem_cache.Items, *d)
	kmem_cache.Lock.Unlock()
}

func New_kmem_cache() (Kmem_cache){
	kmem_cache := Kmem_cache{}
	kmem_cache.Items = make([]Dentry, 10)//make(map[Dentry]bool)
	kmem_cache.Lock = sync.Mutex{}

	return kmem_cache
}
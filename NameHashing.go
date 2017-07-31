package dcache2

import (
	"hash/fnv"
	"fmt"
)

func getLastComponentLengthAndHashFromPath(name string) (uint32, uint32) { //equivalent of hash_name
	var hash uint32 = init_name_hash()
	var length uint32 = 0
	var c uint8 = name[0]

	shouldContinue := true

	for shouldContinue {
		length++
		hash = partial_name_hash(uint32(c), hash)
		c = name[length]
		shouldContinue = (c != 0 && c != uint8('/'))
	}
	return hash, length
}

func init_name_hash() uint32 { //basically salt
	return 0
}

func partial_name_hash(c uint32, prevhash uint32) uint32 {
	return (prevhash + (c<<4) + (c >> 4)) * 11
}

func hashlen_string(salt string, name string) (uint32, uint32) { //returns hash, length
	return hashMeString(salt + name), uint32(len(name))
}

func hash_string(salt string, name string) (uint32) { //returns hash, length
	return hashMeString(salt + name)
}

func hashMeString(data string) uint32 { //TODO: OPTIMIZE
	hasher := fnv.New32a()
	hasher.Write([]byte(data))
	out := hasher.Sum32()
	return uint32(out)
}

func hashMe(data interface{}) uint32 { //TODO: OPTIMIZE
	hasher := fnv.New32a()
	serialized := fmt.Sprintf("%v", data)
	hasher.Write([]byte(serialized))
	out := hasher.Sum32()
	return uint32(out)
}
package btc

import (
	"fmt"
	"github.com/ontio/multi-chain/common"
	"math/big"
	"sort"
)

type BtcProof struct {
	Tx           []byte
	Proof        []byte
	Height       uint32
	BlocksToWait uint64
}

func (this *BtcProof) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Tx)
	sink.WriteVarBytes(this.Proof)
	sink.WriteUint32(this.Height)
	sink.WriteUint64(this.BlocksToWait)
}

func (this *BtcProof) Deserialization(source *common.ZeroCopySource) error {
	tx, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcProof deserialize tx error")
	}
	proof, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcProof deserialize proof error")
	}
	height, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("BtcProof deserialize height error")
	}
	blocksToWait, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("BtcProof deserialize blocksToWait error:")
	}

	this.Tx = tx
	this.Proof = proof
	this.Height = uint32(height)
	this.BlocksToWait = blocksToWait
	return nil
}

type Utxos struct {
	Utxos []*Utxo
}

func (this *Utxos) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.Utxos)))
	for _, v := range this.Utxos {
		v.Serialization(sink)
	}
}

func (this *Utxos) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("utils.DecodeVarUint, deserialize Utxos length error")
	}
	utxos := make([]*Utxo, 0)
	for i := 0; uint64(i) < n; i++ {
		utxo := new(Utxo)
		if err := utxo.Deserialization(source); err != nil {
			return fmt.Errorf("deserialize utxo error: %v", err)
		}
		utxos = append(utxos, utxo)
	}

	this.Utxos = utxos
	return nil
}

func (this *Utxos) Len() int {
	return len(this.Utxos)
}

func (this *Utxos) Less(i, j int) bool {
	vai := new(big.Int).Mul(big.NewInt(int64(this.Utxos[i].Value)), big.NewInt(int64(this.Utxos[i].Confs)))
	vaj := new(big.Int).Mul(big.NewInt(int64(this.Utxos[j].Value)), big.NewInt(int64(this.Utxos[j].Confs)))
	return vai.Cmp(vaj) == -1
}

func (this *Utxos) Swap(i, j int) {
	temp := this.Utxos[i]
	this.Utxos[i] = this.Utxos[j]
	this.Utxos[j] = temp
}

type Utxo struct {
	// Previous txid and output index
	Op *OutPoint

	// Block height where this tx was confirmed, 0 for unconfirmed
	AtHeight uint32

	// The higher the better
	Value uint64

	// Output script
	ScriptPubkey []byte

	// no need to save
	Confs uint32
}

func (this *Utxo) Serialization(sink *common.ZeroCopySink) {
	this.Op.Serialization(sink)
	sink.WriteUint32(this.AtHeight)
	sink.WriteUint64(this.Value)
	sink.WriteVarBytes(this.ScriptPubkey)
}

func (this *Utxo) Deserialization(source *common.ZeroCopySource) error {
	op := new(OutPoint)
	err := op.Deserialization(source)
	if err != nil {
		return fmt.Errorf("Utxo deserialize OutPoint error:%s", err)
	}
	atHeight, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("OutPoint deserialize atHeight error")
	}
	value, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("OutPoint deserialize value error")
	}
	scriptPubkey, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("OutPoint deserialize scriptPubkey error")
	}

	this.Op = op
	this.AtHeight = atHeight
	this.Value = value
	this.ScriptPubkey = scriptPubkey
	return nil
}

type OutPoint struct {
	Hash  []byte
	Index uint32
}

func (this *OutPoint) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.Hash)
	sink.WriteUint32(this.Index)
}

func (this *OutPoint) Deserialization(source *common.ZeroCopySource) error {
	hash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("OutPoint deserialize hash error")
	}
	index, eof := source.NextUint32()
	if eof {
		return fmt.Errorf("OutPoint deserialize height error")
	}

	this.Hash = hash
	this.Index = index
	return nil
}

type MultiSignInfo struct {
	MultiSignInfo map[string][][]byte
}

func (this *MultiSignInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(uint64(len(this.MultiSignInfo)))
	var MultiSignInfoList []string
	for k := range this.MultiSignInfo {
		MultiSignInfoList = append(MultiSignInfoList, k)
	}
	sort.SliceStable(MultiSignInfoList, func(i, j int) bool {
		return MultiSignInfoList[i] > MultiSignInfoList[j]
	})
	for _, k := range MultiSignInfoList {
		sink.WriteString(k)
		v := this.MultiSignInfo[k]
		sink.WriteUint64(uint64(len(v)))
		for _, b := range v {
			sink.WriteVarBytes(b)
		}
	}
}

func (this *MultiSignInfo) Deserialization(source *common.ZeroCopySource) error {
	n, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("MultiSignInfo deserialize MultiSignInfo length error")
	}
	multiSignInfo := make(map[string][][]byte)
	for i := 0; uint64(i) < n; i++ {
		k, eof := source.NextString()
		if eof {
			return fmt.Errorf("MultiSignInfo deserialize public key error")
		}
		m, eof := source.NextUint64()
		if eof {
			return fmt.Errorf("MultiSignInfo deserialize MultiSignItem length error")
		}
		multiSignItem := make([][]byte, 0)
		for j := 0; uint64(j) < m; j++ {
			b, eof := source.NextVarBytes()
			if eof {
				return fmt.Errorf("MultiSignInfo deserialize []byte error")
			}
			multiSignItem = append(multiSignItem, b)
		}
		multiSignInfo[k] = multiSignItem
	}
	this.MultiSignInfo = multiSignInfo
	return nil
}

type Args struct {
	ToChainID         uint64
	Fee               int64
	ToContractAddress []byte
	Address           []byte
}

func (this *Args) Serialization(sink *common.ZeroCopySink) {
	sink.WriteUint64(this.ToChainID)
	sink.WriteInt64(this.Fee)
	sink.WriteVarBytes(this.ToContractAddress)
	sink.WriteVarBytes(this.Address)
}

func (this *Args) Deserialization(source *common.ZeroCopySource) error {
	toChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("Args deserialize toChainID error")
	}
	fee, eof := source.NextInt64()
	if eof {
		return fmt.Errorf("Args deserialize fee error")
	}
	toContractAddress, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args deserialize toContractAddress error")
	}
	address, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("Args deserialize address error")
	}

	this.ToChainID = toChainID
	this.Fee = fee
	this.ToContractAddress = toContractAddress
	this.Address = address
	return nil
}

type BtcFromInfo struct {
	FromTxHash  []byte
	FromChainID uint64
}

func (this *BtcFromInfo) Serialization(sink *common.ZeroCopySink) {
	sink.WriteVarBytes(this.FromTxHash)
	sink.WriteUint64(this.FromChainID)
}

func (this *BtcFromInfo) Deserialization(source *common.ZeroCopySource) error {
	fromTxHash, eof := source.NextVarBytes()
	if eof {
		return fmt.Errorf("BtcProof deserialize fromTxHash error")
	}
	fromChainID, eof := source.NextUint64()
	if eof {
		return fmt.Errorf("BtcProof deserialize fromChainID error:")
	}

	this.FromTxHash = fromTxHash
	this.FromChainID = fromChainID
	return nil
}

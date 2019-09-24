package raft
//
//type Index struct {
//	node			*Node
//	metas			[]*Meta
//	ret				uint64
//}
//func newIndex(node *Node) *Index {
//	i:=&Index{
//		node:node,
//		metas:make([]*Meta,0),
//	}
//	return i
//}
//func (i *Index) appendMetas(metas []*Meta) {
//	i.metas=append(i.metas,metas...)
//	data,_:=i.Encode(metas)
//	i.append(data)
//}
//func (i *Index) checkIndex(index uint64)bool {
//	if index==0{
//		return false
//	}
//	length:=len(i.metas)
//	if length==0{
//		return false
//	}
//	if index<i.metas[0].Index||index>i.metas[length-1].Index{
//		return false
//	}
//	return true
//}
//func (i *Index) lookup(index uint64)*Meta {
//	if !i.checkIndex(index){
//		return nil
//	}
//	for j:=len(i.metas)-1;j>=0;j--{
//		meta:= i.metas[j]
//		if meta.Index==index{
//			return meta
//		}
//	}
//	return nil
//}
//func (i *Index) lookupLast(index uint64)*Meta {
//	if !i.checkIndex(index){
//		return nil
//	}
//	length:=len(i.metas)
//	if length<2{
//		return nil
//	}
//	var cur int
//	for j:=length-1;j>=0;j--{
//		meta:= i.metas[j]
//		if meta.Index==index{
//			cur=j
//			break
//		}
//	}
//	if cur<1{
//		return nil
//	}
//	return i.metas[cur-1]
//}
//func (i *Index) lookupNext(index uint64)*Meta {
//	if !i.checkIndex(index){
//		return nil
//	}
//	length:=len(i.metas)
//	if length<2{
//		return nil
//	}
//	var cur int
//	for j:=length-1;j>=0;j--{
//		meta:= i.metas[j]
//		if meta.Index==index{
//			cur=j
//			break
//		}
//	}
//	if cur>length-2{
//		return nil
//	}
//	return i.metas[cur+1]
//}
//func (i *Index) deleteAfter(index uint64){
//	length:=len(i.metas)
//	if length==0{
//		return
//	}
//	if index<=1&&length>0{
//		i.metas=i.metas[length:]
//		return
//	}
//	for j:=length-1;j>=0;j--{
//		meta:= i.metas[j]
//		if meta.Index==index{
//			i.metas=i.metas[:j]
//			break
//		}
//	}
//	i.save()
//	Tracef("Index.deleteAfter %s delete %d and after",i.node.address, index)
//}
//func (i *Index) copyAfter(index uint64,max int)(metas []*Meta) {
//	length:=len(i.metas)
//	if length==0{
//		return
//	}
//	if index<=1&&length>0{
//		if max<=length{
//			metas=i.metas[0:max]
//		}else {
//			metas=i.metas[:]
//		}
//		return
//	}
//	for j:=len(i.metas)-1;j>=0;j--{
//		meta:= i.metas[j]
//		if meta.Index==index{
//			if j+max<=len(i.metas){
//				metas=i.metas[j:j+max]
//			}else {
//				metas=i.metas[j:]
//			}
//			break
//		}
//	}
//	return
//}
//func (i *Index) copyRange(startIndex uint64,endIndex uint64) []*Meta{
//	if startIndex>endIndex{
//		return nil
//	}
//	length:=len(i.metas)
//	if length==0{
//		return nil
//	}
//	if endIndex<i.metas[0].Index||startIndex>i.metas[length-1].Index{
//		return nil
//	}
//	var metas []*Meta
//	for j:=0;j<len(i.metas);j++{
//		meta:=i.metas[j]
//		if meta.Index<startIndex{
//			continue
//		}else if meta.Index>endIndex{
//			break
//		}
//		metas=append(metas, meta)
//	}
//	return metas
//}
//func (i *Index) append(b []byte) {
//	i.node.storage.SeekWrite(DefaultIndex,i.ret,b)
//	i.ret+=uint64(len(b))
//}
//func (i *Index) save() {
//	b,_:=i.Encode(i.metas)
//	i.node.storage.SafeOverWrite(DefaultIndex,b)
//	i.ret=uint64(len(b))
//}
//func (i *Index) recover() error {
//	if !i.node.storage.Exists(DefaultIndex){
//		return nil
//	}
//	b, err := i.node.storage.Load(DefaultIndex)
//	if err != nil {
//		return err
//	}
//	if err = i.Decode(b); err != nil {
//		return err
//	}
//	i.ret=uint64(len(i.metas)*32)
//	return nil
//}
//func (i *Index)Decode(data []byte)error  {
//	cnt:=len(data)/32
//	for j:=0;j<cnt;j++{
//		b:=data[j*32:j*32+32]
//		meta:=&Meta{}
//		meta.Decode(b)
//		i.metas=append(i.metas, meta)
//	}
//	Tracef("Index.Decode %s Metas length %d",i.node.address,cnt)
//	return nil
//}
//func (i *Index)Encode(metas []*Meta)([]byte,error)  {
//	var data=[]byte{}
//	for j:=0;j<len(metas);j++{
//		b:=metas[j].Encode()
//		data=append(data,b...)
//	}
//	return data,nil
//}

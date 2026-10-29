export function paginate<T>(items:T[], requestedPage:number, pageSize=10) {
  const pageCount=Math.max(1,Math.ceil(items.length/pageSize))
  const page=Math.max(1,Math.min(requestedPage,pageCount))
  const offset=(page-1)*pageSize
  return {items:items.slice(offset,offset+pageSize),page,pageCount,start:items.length?offset+1:0,end:Math.min(offset+pageSize,items.length),total:items.length}
}

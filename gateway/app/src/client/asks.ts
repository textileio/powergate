import {grpc} from '@improbable-eng/grpc-web'
import {APIClient} from '../_proto/ask_pb_service'
import {GetRequest, QueryRequest, Index, Query, StorageAsk} from '../_proto/ask_pb'

export default class Asks {

  private client: APIClient

  constructor(serviceHost: string, options?: grpc.RpcOptions) {
    this.client = new APIClient(serviceHost, options)
  }

  get() {
    return new Promise<Index.AsObject>((resolve, reject) => {
      this.client.get(new GetRequest(), (error, resp) => {
        if (error) {
          reject(error)
        } else {
          const val = resp?.getIndex()?.toObject()
          if (!val) {
            reject(new Error('no response object'))
          } else {
            resolve(val)
          }
        }
      })
    })
  }

  query(query: Query) {
    const r = new QueryRequest()
    r.setQuery(query)
    return new Promise<StorageAsk.AsObject[]>((resolve, reject) => {
      this.client.query(r, (error, resp) => {
        if (error) {
          reject(error)
        } else {
          const val = resp?.toObject()?.asksList
          if (!val) {
            reject(new Error('no response object'))
          } else {
            resolve(val)
          }
        }
      })
    })
  }
}
import {grpc} from '@improbable-eng/grpc-web'
import {APIClient, API} from '../_proto/miner_pb_service'
import {GetRequest, GetReply, Index} from '../_proto/miner_pb'

export default class Miners {

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
}

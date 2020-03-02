import {grpc} from '@improbable-eng/grpc-web'
import {APIClient} from '../_proto/reputation_pb_service'
import {GetTopMinersRequest, MinerScore} from '../_proto/reputation_pb'

export default class Reputation {

  private client: APIClient

  constructor(serviceHost: string, options?: grpc.RpcOptions) {
    this.client = new APIClient(serviceHost, options)
  }

  getTopMiners(limit: number = 20) {
    return new Promise<MinerScore.AsObject[]>((resolve, reject) => {
      const request = new GetTopMinersRequest()
      request.setLimit(limit)
      this.client.getTopMiners(request, (error, resp) => {
        if (error) {
          reject(error)
        } else {
          const val = resp?.toObject()?.topminersList
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

import MultiView from '@/components/MultiView';
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import {CatalogState, User} from '../types';
const getCatalogs = async (type: string) => {
    const session = await getServerSession(options);    
    const symphonyApi = process.env.SYMPHONY_API;
    const userObj: User | undefined = session?.user?? undefined;
    const res = await fetch( `${symphonyApi}catalogs/graph?template=${type}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
      }
    });
    const data = await res.json();
    return data;
  }
async function AssetsPage() {
    const [catalogs, configs] =  await Promise.all([getCatalogs('asset-trees'), getCatalogs('config-chains')]);

    const params = {
        type: 'assets',
        menuItems: [           
        ],
        views: ['cards', 'table'],
        items: catalogs,
        refItems: [],
        columns: [{
          name: 'configs',
          data: configs
        }, {
          name: 'solutions'
        }, {
          name: 'instances'
        }, {
          name: 'targets'
        }],
    }
    return (
        <div>
            <MultiView params={params}  />
        </div>
    );
}

export default AssetsPage;
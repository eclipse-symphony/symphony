import MultiView from '@/components/MultiView';
import { getServerSession } from 'next-auth';
import { options } from '../api/auth/[...nextauth]/options';
import {CatalogVersionState, User} from '../types';
const getCatalogVersions = async (type: string) => {
    const session = await getServerSession(options);    
    const symphonyApi = process.env.SYMPHONY_API;
    const userObj: User | undefined = session?.user?? undefined;
    const res = await fetch( `${symphonyApi}catalogversions/graph?template=${type}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${userObj?.accessToken}`,
      }
    });
    const data = await res.json();    
    return data;
  }
async function AssetsPage() {
    const [catalogversions, configs] =  await Promise.all([getCatalogVersions('asset-trees'), getCatalogVersions('config-chains')]);

    const params = {
        type: 'assets',
        menuItems: [           
        ],
        views: ['cards', 'table'],
        items: catalogversions,
        refItems: [],
        columns: [{
          name: 'configs',
          data: configs
        }, {
          name: 'solutionversions'
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
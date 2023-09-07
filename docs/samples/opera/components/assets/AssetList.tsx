import { CatalogState } from '../../app/types';
import AssetCard from './AssetCard';
interface AssetListProps {
    catalogs: CatalogState[];
}
function AssetList(props: AssetListProps) {
    const { catalogs } = props;
    //create a map of catalogs
    const references: any = {};
    for (const [_, cats] of Object.entries(catalogs)) {
        cats.forEach((catalog: CatalogState) => {
            references[catalog.spec.name] = catalog;
        });
    }
    const mergedCatalogs = [];
    for (const [_, cats] of Object.entries(catalogs)) {
        mergedCatalogs.push(...cats);
    }

    return (
        <div className='sitelist'>            
            {mergedCatalogs.map((catalog: CatalogState) =>  
            <AssetCard catalog={catalog} refCatalog={catalog.spec.metadata?.['override']? references[catalog.spec.metadata['override']]: null}/>)}
        </div>
    );
}
export default AssetList;
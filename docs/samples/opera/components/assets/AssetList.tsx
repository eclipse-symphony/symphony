import { CatalogState } from '../../app/types';
import AssetCard from './AssetCard';

interface AssetListProps {
    catalogs: CatalogState[];
}

function AssetList({ catalogs }: AssetListProps) {
    // Create a map of catalogs by name for easy reference
    const references: Record<string, CatalogState> = {};
    catalogs.forEach((catalog) => {
        references[catalog.spec.name] = catalog;
    });

    // If you want to merge catalogs or perform other operations, you can do it directly with the array
    // Assuming mergedCatalogs is supposed to be the same as catalogs in this simplified correction
    const mergedCatalogs = [...catalogs]; // This creates a shallow copy if needed, or directly use catalogs

    return (
        <div className='sitelist'>
            {mergedCatalogs.map((catalog) => (
                <AssetCard 
                    key={catalog.spec.name} // It's a good practice to provide a unique key for each child in a list
                    catalog={catalog} 
                    refCatalog={catalog.spec.metadata?.['override'] ? references[catalog.spec.metadata['override']] : null}
                />
            ))}
        </div>
    );
}

export default AssetList;

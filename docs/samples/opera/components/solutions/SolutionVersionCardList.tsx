/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

import {SolutionVersionState} from '../../app/types';
import SolutionVersionCard from './SolutionVersionCard';

interface SolutionVersionCardListProps {
    solutionversions: SolutionVersionState[];
    activations?: any[];
}

function SolutionVersionCardList(props: SolutionVersionCardListProps) {
    const { solutionversions } = props;
    if (!solutionversions) {
        return (<div>No data</div>);
    }

    return (
        <div className='sitelist'>            
            {solutionversions.map((solutionversion: any) =>  {                
                return <SolutionVersionCard solutionversion={solutionversion} />;
            })}
        </div>
    );
}

export default SolutionVersionCardList;
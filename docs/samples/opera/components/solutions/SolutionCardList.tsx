/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

import {SolutionState} from '../../app/types';
import SolutionCard from './SolutionCard';

interface SolutionCardListProps {
    solutions: SolutionState[];
    activations?: any[];
}

function SolutionCardList(props: SolutionCardListProps) {
    const { solutions } = props;
    if (!solutions) {
        return (<div>No data</div>);
    }

    return (
        <div className='sitelist'>            
            {solutions.map((solution: any) =>  {                
                return <SolutionCard solution={solution} />;
            })}
        </div>
    );
}

export default SolutionCardList;
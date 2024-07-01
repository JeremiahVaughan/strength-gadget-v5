import { render } from '@testing-library/react';

import MeasurementMiles from './measurement-miles';

describe('MeasurementMiles', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<MeasurementMiles />);
    expect(baseElement).toBeTruthy();
  });
});

import { render } from '@testing-library/react';

import ToolBar from './tool-bar';

describe('ToolBar', () => {
  it('should render successfully', () => {
    const { baseElement } = render(<ToolBar />);
    expect(baseElement).toBeTruthy();
  });
});

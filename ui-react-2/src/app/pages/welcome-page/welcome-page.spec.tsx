import {render} from '@testing-library/react';

import WelcomePage from './welcome-page';
import {BrowserRouter} from "react-router-dom";

describe('WelcomePage', () => {
  it('should render successfully', () => {
    const { baseElement } = render(
        <BrowserRouter>
          <WelcomePage />
        </BrowserRouter>
    );
    expect(baseElement).toBeTruthy();
  });
});
